package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migration/*.sql
var migrationFS embed.FS

type DB struct {
	db     *sql.DB
	ctx    context.Context
	cancel func()

	DSN string
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
	}
	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

func (db *DB) Open() (err error) {
	// validate the a dsn is provided
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	// make the parent directory if using a file
	if db.DSN != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(db.DSN), 0700); err != nil {
			return err
		}
	}

	// connect to the database
	if db.db, err = sql.Open("sqlite3", db.DSN); err != nil {
		return err
	}

	// configure the database
	if err = db.configure(); err != nil {
		return err
	}

	// migrate the database
	if err = db.migrate(); err != nil {
		return err
	}

	return nil
}

func (db *DB) configure() error {
	// enable wal
	if _, err := db.db.Exec(`PRAGMA journal_mode = wal;`); err != nil {
		return fmt.Errorf("enable wal: %w", err)
	}

	// enable foreign keys
	if _, err := db.db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	// configure busy timeout
	if _, err := db.db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		return fmt.Errorf("set busy_timeout: %w", err)
	}

	// configure synchronous mode
	// note: trades a little durability for much faster writes
	if _, err := db.db.Exec(`PRAGMA synchronous = NORMAL;`); err != nil {
		return fmt.Errorf("set synchronous: %w", err)
	}

	// configure temp store in memory (instead of disk)
	if _, err := db.db.Exec(`PRAGMA temp_store = MEMORY;`); err != nil {
		return fmt.Errorf("set temp_store: %w", err)
	}

	return nil
}

func (db *DB) migrate() error {
	// ensure the 'migrations' table exists
	if _, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`); err != nil {
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	names, err := fs.Glob(migrationFS, "migration/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	for _, name := range names {
		if err := db.migrateFile(name); err != nil {
			return fmt.Errorf("migration error: name=%q err=%w", name, err)
		}
	}

	return nil
}

func (db *DB) migrateFile(name string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check if the migration has already been run
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = ?`, name).Scan(&n); err != nil {
		return err
	} else if n != 0 {
		return nil // already run migration, skip
	}

	// read and execute migration file.
	if buf, err := fs.ReadFile(migrationFS, name); err != nil {
		return err
	} else if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	// insert record into migrations to prevent re-running migration.
	if _, err := tx.Exec(`INSERT INTO migrations (name) VALUES (?)`, name); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) Close() error {
	// cancel background context.
	db.cancel()

	// close database.
	if db.db != nil {
		return db.db.Close()
	}

	return nil
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:  tx,
		db:  db,
		now: time.Now().UTC().Truncate(time.Second),
	}, nil
}

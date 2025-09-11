package sqlite

import (
	"database/sql"
	"time"
)

type Tx struct {
	*sql.Tx
	db  *DB
	now time.Time
}

package sqlite

import (
	"context"

	"github.com/heyjorgedev/suss"
)

type ShortURLService struct {
	db *DB
}

func NewShortURLService(db *DB) *ShortURLService {
	return &ShortURLService{
		db: db,
	}
}

func (s *ShortURLService) Create(ctx context.Context, shortURL *suss.ShortURL) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := shortUrlCreate(ctx, tx, shortURL); err != nil {
		return err
	}

	return tx.Commit()
}

func shortUrlCreate(ctx context.Context, tx *Tx, s *suss.ShortURL) error {
	// generate a unique slug
	slug, err := shortUrlGenerateSlug(tx)
	if err != nil {
		return err
	}
	s.Slug = slug

	// generate secret key
	secretKey, err := shortUrlGenerateSecretKey(tx)
	if err != nil {
		return err
	}
	s.SecretKey = secretKey

	// set created and updated at
	s.CreatedAt = tx.now
	s.UpdatedAt = s.CreatedAt

	// validate the short url
	if err := s.Validate(); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO short_urls (slug, long_url, secret_key, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, s.Slug, s.LongURL, s.SecretKey, (*NullTime)(&s.CreatedAt), (*NullTime)(&s.UpdatedAt))
	if err != nil {
		return err
	}

	// Read back new dial ID into caller argument.
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	s.ID = int(id)

	return nil
}

func shortUrlGenerateSlug(tx *Tx) (string, error) {
	return "", nil
}

func shortUrlGenerateSecretKey(tx *Tx) (string, error) {
	return "", nil
}

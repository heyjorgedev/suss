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

	if err := shortUrlCreate(tx, shortURL); err != nil {
		return err
	}

	return tx.Commit()
}

func shortUrlCreate(tx *Tx, s *suss.ShortURL) error {
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

	// validate the short url
	if err := s.Validate(); err != nil {
		return err
	}

	return nil
}

func shortUrlGenerateSlug(tx *Tx) (string, error) {
	return "", nil
}

func shortUrlGenerateSecretKey(tx *Tx) (string, error) {
	return "", nil
}

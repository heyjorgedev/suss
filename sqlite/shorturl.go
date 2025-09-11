package sqlite

import "github.com/heyjorgedev/suss"

type ShortURLService struct {
	db *DB
}

func NewShortURLService(db *DB) *ShortURLService {
	return &ShortURLService{
		db: db,
	}
}

func (s *ShortURLService) Create(shortURL *suss.ShortURL) error {
	return nil
}

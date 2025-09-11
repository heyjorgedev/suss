package suss

import "context"

type ShortURL struct {
	ID        int    `json:"id"`
	Slug      string `json:"slug"`
	LongURL   string `json:"long_url"`
	SecretKey string `json:"secret_key"`
}

func (s *ShortURL) Validate() error {
	return nil
}

type ShortURLService interface {
	Create(ctx context.Context, shortURL *ShortURL) error
}

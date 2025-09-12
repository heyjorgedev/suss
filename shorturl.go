package suss

import (
	"context"
	"time"
)

type ShortURL struct {
	ID        int    `json:"id"`
	Slug      string `json:"slug"`
	LongURL   string `json:"long_url"`
	SecretKey string `json:"secret_key"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *ShortURL) Validate() error {
	return nil
}

type ShortURLFilter struct {
	Slug *string `json:"slug"`
}

type ShortURLService interface {
	FindShortUrls(ctx context.Context, filter ShortURLFilter) ([]*ShortURL, int, error)
	FindDialBySlug(ctx context.Context, slug string) (*ShortURL, error)
	Create(ctx context.Context, shortURL *ShortURL) error
}

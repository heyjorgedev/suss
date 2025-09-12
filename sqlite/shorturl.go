package sqlite

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math/big"
	"strings"

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

func (s *ShortURLService) FindShortUrls(ctx context.Context, filter suss.ShortURLFilter) ([]*suss.ShortURL, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	// Fetch list of matching dial objects.
	shortUrls, n, err := findShortUrls(ctx, tx, filter)
	if err != nil {
		return shortUrls, n, err
	}

	return shortUrls, n, nil
}

func (s *ShortURLService) FindDialBySlug(ctx context.Context, slug string) (*suss.ShortURL, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	shortUrl, err := findShortUrlBySlug(ctx, tx, slug)
	if err != nil {
		return nil, err
	}

	return shortUrl, nil
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

const slugAlphabet = "abcdefghijkmnopqrstuvwxyz" + "23456789" // avoid 0's and o's, 1's l's - for less ambiguity
const slugLength = 6                                          // adjust length depending on your collision risk tolerance

func generateRandomSlug() (string, error) {
	result := make([]byte, slugLength)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(slugAlphabet))))
		if err != nil {
			return "", err
		}
		result[i] = slugAlphabet[n.Int64()]
	}
	return string(result), nil
}

func shortUrlGenerateSlug(tx *Tx) (string, error) {
	const maxAttempts = 10 // avoid infinite loops
	for i := 0; i < maxAttempts; i++ {
		// generate a random slug
		slug, err := generateRandomSlug()
		if err != nil {
			return "", err
		}

		// check DB if slug already exists
		var exists bool
		err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM short_urls WHERE slug = ?)`, slug).Scan(&exists)
		if err != nil {
			return "", err
		}

		// if not exists, return it
		if !exists {
			return slug, nil
		}
	}

	return "", errors.New("could not generate unique slug after several attempts")
}

func shortUrlGenerateSecretKey(tx *Tx) (string, error) {
	// 32 bytes random secret -> base64 encoded
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}

	// Use URL-safe base64
	return base64.RawURLEncoding.EncodeToString(secret), nil
}

func findShortUrls(ctx context.Context, tx *Tx, filter suss.ShortURLFilter) ([]*suss.ShortURL, int, error) {
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.Slug; v != nil {
		where, args = append(where, "slug = ?"), append(args, *v)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, slug, long_url, secret_key, created_at, updated_at, COUNT(*) OVER()
		FROM short_urls
		WHERE `+strings.Join(where, " AND "), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	n := 0
	shortUrls := make([]*suss.ShortURL, 0)
	for rows.Next() {
		var shortUrl suss.ShortURL
		if err := rows.Scan(
			&shortUrl.ID,
			&shortUrl.Slug,
			&shortUrl.LongURL,
			&shortUrl.SecretKey,
			(*NullTime)(&shortUrl.CreatedAt),
			(*NullTime)(&shortUrl.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, err
		}
		shortUrls = append(shortUrls, &shortUrl)
	}

	return shortUrls, 0, nil
}

func findShortUrlBySlug(ctx context.Context, tx *Tx, slug string) (*suss.ShortURL, error) {
	shortUrls, _, err := findShortUrls(ctx, tx, suss.ShortURLFilter{Slug: &slug})
	if err != nil {
		return nil, err
	}
	if len(shortUrls) == 0 {
		return nil, &suss.Error{Code: suss.ENOTFOUND, Message: "Short Url not found."}
	}

	return shortUrls[0], nil
}

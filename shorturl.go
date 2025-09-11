package suss

type ShortURL struct {
	ID        int    `json:"id"`
	Slug      string `json:"slug"`
	LongURL   string `json:"long_url"`
	SecretKey string `json:"secret_key"`
}

type ShortURLService interface {
	Create(shortURL *ShortURL) error
}

CREATE TABLE short_urls (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	slug TEXT UNIQUE NOT NULL,
	long_url TEXT NOT NULL,
	secret_key TEXT NOT NULL,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL
);


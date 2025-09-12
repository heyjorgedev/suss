package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/heyjorgedev/suss"
	"github.com/heyjorgedev/suss/http"
	"github.com/heyjorgedev/suss/sqlite"
)

func main() {
	suss.Version = "0.0.1"
	suss.Commit = "1234567890"

	// setup signal handlers
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	p := NewProgram()

	if err := p.Run(ctx); err != nil {
		p.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	<-ctx.Done()

	if err := p.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}

type Config struct {
	DB struct {
		DSN string
	}
	Hostname string
	Port     int
}

func DefaultConfig() *Config {
	config := &Config{}

	// database
	config.DB.DSN = ":memory:"

	// http
	config.Hostname = "0.0.0.0"
	config.Port = 8080

	return config
}

func GetConfigFromEnv() (*Config, error) {
	config := DefaultConfig()

	dsn := os.Getenv("DB_DSN")
	if dsn != "" {
		config.DB.DSN = dsn
	}

	hostname := os.Getenv("HOSTNAME")
	if hostname != "" {
		config.Hostname = hostname
	}

	port := os.Getenv("PORT")
	if port != "" {
		portInt, err := strconv.Atoi(port)
		if err != nil {
			return config, fmt.Errorf("invalid port: %w", err)
		}
		config.Port = portInt
	}

	return config, nil
}

type Program struct {
	// configuration
	Config *Config

	// sqlite database
	DB *sqlite.DB

	// http server
	HTTPServer *http.Server

	// services
	ShortURLService suss.ShortURLService
}

func NewProgram() *Program {
	return &Program{
		Config:     DefaultConfig(),
		DB:         sqlite.NewDB(":memory:"),
		HTTPServer: http.NewServer(),
	}
}

func (p *Program) Run(ctx context.Context) error {
	// load the config from the environment
	config, err := GetConfigFromEnv()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}
	p.Config = config

	// open the database, configure it and migrate to latest version
	p.DB.DSN = p.Config.DB.DSN
	if err := p.DB.Open(); err != nil {
		return fmt.Errorf("cannot open db: %w", err)
	}

	// initialize services
	p.ShortURLService = sqlite.NewShortURLService(p.DB)

	// bind services to http server
	p.HTTPServer.ShortURLService = p.ShortURLService

	// configure http server
	p.HTTPServer.Addr = fmt.Sprintf("%s:%d", p.Config.Hostname, p.Config.Port)

	// start the http server
	if err := p.HTTPServer.Open(); err != nil {
		return err
	}

	return nil
}

func (p *Program) Close() error {
	// close the database
	if p.DB != nil {
		if err := p.DB.Close(); err != nil {
			return fmt.Errorf("cannot close db: %w", err)
		}
	}

	return nil
}

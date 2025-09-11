package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

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

type Program struct {
	// sqlite database
	DB *sqlite.DB

	// http server
	HTTPServer *http.Server
}

func NewProgram() *Program {
	return &Program{
		DB:         sqlite.NewDB(":memory:"),
		HTTPServer: http.NewServer(),
	}
}

func (p *Program) Run(ctx context.Context) error {
	// open the database, configure it and migrate to latest version
	if err := p.DB.Open(); err != nil {
		return fmt.Errorf("cannot open db: %w", err)
	}

	// configure http server
	p.HTTPServer.Addr = ":8080"

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

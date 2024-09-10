package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/jwtmanager"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/server"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/store"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		fmt.Printf("unexpected error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("gracefully stopped")
}

func run(ctx context.Context) error {
	pgxCfg, err := pgxpool.ParseConfig("postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable")
	if err != nil {
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	storage := store.New(pool)
	if err := storage.ApplyMigrations(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	s := server.New(jwtmanager.New([]byte("")), storage)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutting down http server", "error", err)
		}
	}()

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

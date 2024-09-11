package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		log.Fatalf("unexpected error: %v", err)
	}

	fmt.Println("gracefully stopped")
}

func run(ctx context.Context) error {
	// databaseURL := os.Getenv("DATABASE_URL")
	// if databaseURL == "" {
	// 	return fmt.Errorf("DATABASE_URL not set")
	// }

	pgxCfg, err := pgxpool.ParseConfig("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
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

	// jwtSecret := os.Getenv("JWT_SECRET")
	// if jwtSecret == "" {
	// 	return fmt.Errorf("JWT_SECRET not set")
	// }

	s := server.New(jwtmanager.New([]byte("jwtSecret")), storage)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			log.Printf("error shutting down http server: %v", err)
		}
	}()

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

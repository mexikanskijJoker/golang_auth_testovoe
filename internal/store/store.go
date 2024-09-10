package store

import (
	"context"
	_ "embed"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenStore interface {
}

type Store struct {
	db *pgxpool.Pool
}

var (
	userTable = goqu.Dialect("postgres").From("users").Prepared(true)
	userCols  = []any{
		"id",
		"email",
	}
)

var (
	refreshTokenTable = goqu.Dialect("postgres").From("refresh_tokens").Prepared(true)
	refreshTokenCols  = []any{
		"payload",
	}
)

var (
	//go:embed 1_init.sql
	migrationFS []byte
)

var _ TokenStore = (*Store)(nil)

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		db: pool,
	}
}

func (s *Store) ApplyMigrations(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, string(migrationFS)); err != nil {
		return err
	}

	return nil
}

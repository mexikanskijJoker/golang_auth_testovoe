package store

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenStore interface {
	CreateUser(email, ip, guid string) error
	CreateRefreshToken(payload string) error
}

type Store struct {
	db *pgxpool.Pool
}

var (
	userTable = goqu.Dialect("postgres").From("users").Prepared(true)
	userCols  = []any{
		"email",
		"uuid",
		"current_ip_sign_in",
		"last_sign_in_at",
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

func (s *Store) CreateUser(email, ip, guid string) error {
	sql, args, err := userTable.
		Insert().
		Rows(goqu.Record{
			"uuid":               guid,
			"email":              email,
			"current_ip_sign_in": ip,
			"last_sign_in_at":    time.Now(),
		}).
		Returning(userCols...).ToSQL()
	if err != nil {
		return fmt.Errorf("create_user: %w", err)
	}

	_, err = s.db.Exec(context.Background(), sql, args...)
	if err != nil {
		return fmt.Errorf("exec create_user: %w", err)
	}

	return nil
}

func (s *Store) CreateRefreshToken(payload string) error {
	sql, args, err := refreshTokenTable.
		Insert().
		Rows(goqu.Record{
			"payload": payload,
		}).
		Returning(refreshTokenCols...).ToSQL()
	if err != nil {
		return fmt.Errorf("create_refresh_token: %w", err)
	}

	_, err = s.db.Exec(context.Background(), sql, args...)
	if err != nil {
		return fmt.Errorf("exec create_refresh_token: %w", err)
	}

	return nil
}

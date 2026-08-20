package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

type Vault struct{}

func Open(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("store: parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("store: connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) CreateVault(ctx context.Context, name string, wrappedKey []byte) (Vault, error) {
	return Vault{}, nil
}

func (s *Store) VaultByName(ctx context.Context, name string) (Vault, error) {
	return Vault{}, nil
}

func (s *Store) ListVaults(ctx context.Context) ([]Vault, error) {
	return []Vault{}, nil
}

func (s *Store) DeleteVault(ctx context.Context, name string) error {
	return nil
}

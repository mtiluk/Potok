package store

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/michaeltukdev/Potok/migrations"
)

func (s *Store) Migrate() error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("store: read migrations: %w", err)
	}

	db := stdlib.OpenDBFromPool(s.pool)
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("store: migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("store: migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("store: apply migrations: %w", err)
	}
	return nil
}

func (s *Store) Version() (version uint, dirty bool, err error) {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return 0, false, fmt.Errorf("store: read migrations: %w", err)
	}

	db := stdlib.OpenDBFromPool(s.pool)
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("store: migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return 0, false, fmt.Errorf("store: migrator: %w", err)
	}

	version, dirty, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("store: read schema version: %w", err)
	}
	return version, dirty, nil
}

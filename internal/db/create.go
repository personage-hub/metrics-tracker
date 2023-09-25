package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const migrationsDir = "./internal/migrations"

type Database struct {
	Conn *pgx.Conn
}

func CreateAndConnect(connString string) (Database, error) {
	var db Database
	connConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return db, err
	}
	ctx := context.TODO()

	db.Conn, err = pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return db, err
	}
	return db, nil
}

func (db *Database) Close() {
	db.Conn.Close(context.TODO())
}

func (db *Database) DoMigrations() error {
	sqlDB := stdlib.OpenDB(*db.Conn.Config())

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not start SQL migration: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("an error occurred while syncing migrations: %v", err)
	}

	return nil
}

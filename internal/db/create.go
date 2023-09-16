package db

import (
	"context"
	"github.com/jackc/pgx/v5"
)

const (
	createGaugeTableSQL = `
		CREATE TABLE IF NOT EXISTS gauges (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		value DOUBLE PRECISION NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW());`

	createCounterTableSQL = `
		CREATE TABLE IF NOT EXISTS counters (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			value BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW());`
)

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

func (db *Database) CreateTables() error {

	if _, err := db.Conn.Exec(context.TODO(), createGaugeTableSQL); err != nil {
		return err
	}

	if _, err := db.Conn.Exec(context.TODO(), createCounterTableSQL); err != nil {
		return err
	}
	return nil
}

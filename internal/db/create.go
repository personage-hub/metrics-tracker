package db

import (
	"context"
	"github.com/jackc/pgx/v5"
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

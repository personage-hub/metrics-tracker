package dumper

import (
	"context"
	"fmt"
	"github.com/personage-hub/metrics-tracker/internal/db"
)

type DBDumper struct {
	DB db.Database
}

func NewDBDumper(database db.Database) *DBDumper {
	return &DBDumper{
		DB: database,
	}
}

func (d *DBDumper) SaveData(gaugeMap map[string]float64, counterMap map[string]int64) error {
	ctx := context.Background()

	tx, err := d.DB.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Вставка данных Gauge
	for name, value := range gaugeMap {
		_, err := tx.Exec(ctx, `INSERT INTO gauges (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`, name, value)
		if err != nil {
			return fmt.Errorf("unable to insert/update gauge: %w", err)
		}
	}

	for name, value := range counterMap {
		_, err := tx.Exec(ctx, `INSERT INTO counters (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`, name, value)
		if err != nil {
			return fmt.Errorf("unable to insert/update counter: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("unable to commit transaction: %w", err)
	}
	return nil
}

func (d *DBDumper) RestoreData() (gaugeMap map[string]float64, counterMap map[string]int64, err error) {
	ctx := context.Background()
	gaugeMap = make(map[string]float64)
	counterMap = make(map[string]int64)

	rows, err := d.DB.Conn.Query(ctx, `SELECT name, value FROM gauges`)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to query gauges: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, nil, fmt.Errorf("unable to scan gauge: %w", err)
		}
		gaugeMap[name] = value
	}

	rows, err = d.DB.Conn.Query(ctx, `SELECT name, value FROM counters`)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to query counters: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, nil, fmt.Errorf("unable to scan counter: %w", err)
		}
		counterMap[name] = value
	}

	return gaugeMap, counterMap, nil
}

package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/personage-hub/metrics-tracker/internal/db"
	"go.uber.org/zap"
	"log"
	"time"
)

type DBStorage struct {
	db     *db.Database
	logger *zap.Logger
}

func NewDBStorage(database *db.Database, logger *zap.Logger) *DBStorage {
	return &DBStorage{
		db:     database,
		logger: logger,
	}
}

func (s *DBStorage) GaugeUpdate(key string, value float64) {
	_, _ = s.db.Conn.Exec(context.Background(),
		`INSERT INTO gauges (name, value, created_at, updated_at)
			VALUES ($1, $2, $3, $3)
			ON CONFLICT (name)
			DO UPDATE SET value = $2, updated_at = $3`,
		key, value, time.Now())

}

func (s *DBStorage) CounterUpdate(key string, value int64) {
	_, _ = s.db.Conn.Exec(context.Background(),
		`INSERT INTO counters (name, value, created_at, updated_at) 
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (name) 
		DO UPDATE SET value = counters.value + $2, updated_at = $3`,
		key, value, time.Now())

}

func (s *DBStorage) GaugeMap() map[string]float64 {
	query := "SELECT name, value FROM gauges"
	rows, err := s.db.Conn.Query(context.Background(), query)
	if err != nil {
		log.Println("Error querying the database:", err)
		return nil
	}
	defer rows.Close()

	gaugeMap := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64

		err = rows.Scan(&name, &value)
		if err != nil {
			log.Println("Error scanning the row:", err)
			continue
		}
		gaugeMap[name] = value
	}

	// проверка на ошибки, возникшие при итерации по строкам
	if err = rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return nil
	}

	return gaugeMap
}

func (s *DBStorage) CounterMap() map[string]int64 {
	query := "SELECT name, value FROM counters"
	rows, err := s.db.Conn.Query(context.Background(), query)
	if err != nil {
		log.Println("Error querying the database:", err)
		return nil
	}
	defer rows.Close()

	counterMap := make(map[string]int64)
	for rows.Next() {
		var name string
		var value int64

		err = rows.Scan(&name, &value)
		if err != nil {
			log.Println("Error scanning the row:", err)
			continue
		}
		counterMap[name] = value
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return nil
	}

	return counterMap
}

func (s *DBStorage) GetGaugeMetric(metricName string) (float64, bool) {
	var value float64

	query := "SELECT value FROM gauges WHERE name = $1"
	err := s.db.Conn.QueryRow(context.Background(), query, metricName).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false // метрика не найдена
		}
		log.Println("Error querying the gauge metric:", err)
		return 0, false
	}

	return value, true
}

func (s *DBStorage) GetCounterMetric(metricName string) (int64, bool) {
	var value int64

	query := "SELECT value FROM counters WHERE name = $1"
	err := s.db.Conn.QueryRow(context.Background(), query, metricName).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false // метрика не найдена
		}
		log.Println("Error querying the counter metric:", err)
		return 0, false
	}

	return value, true
}

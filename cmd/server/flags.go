package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress string
	FlagLogLevel  string
	StoreInterval int64
	FileStorage   string
	Restore       bool
}

func isValidPath(path string) bool {
	_, err := os.Open(path)
	if err != nil {
		return os.IsNotExist(err)
	}
	return true
}

func parseFlags() Config {
	var config Config
	flag.StringVar(
		&config.ServerAddress,
		"a",
		"localhost:8080",
		"Address of the HTTP server endpoint",
	)
	flag.StringVar(
		&config.FlagLogLevel,
		"l",
		"info",
		"Logging level",
	)
	flag.Int64Var(
		&config.StoreInterval,
		"i",
		300,
		"The time interval in seconds after which the current data are saved to disk "+
			"(a value of 0 makes the write synchronous)",
	)
	flag.StringVar(
		&config.FileStorage,
		"f",
		"/tmp/metrics-db.json",
		"The full filename where the current values are saved (an empty value disables the disk writing function)",
	)
	flag.BoolVar(
		&config.Restore,
		"r",
		true,
		"A boolean setting that dictates if the server should load values saved earlier from storage upon startup ",
	)

	if envValue := os.Getenv("ADDRESS"); envValue != "" {
		config.ServerAddress = envValue
	}
	flag.Parse()
	if envValue := os.Getenv("STORE_INTERVAL"); envValue != "" {
		val, err := strconv.ParseInt(envValue, 10, 64)
		if err == nil {
			_ = fmt.Errorf("failed parse interval value: %w", err)
		}
		config.StoreInterval = val
	}
	if envValue := os.Getenv("FILE_STORAGE_PATH"); envValue != "" {
		if isValidPath(envValue) {
			config.FileStorage = envValue
		}
	}
	if envValue := os.Getenv("RESTORE"); envValue != "" {
		boolValue, err := strconv.ParseBool(envValue)
		if err != nil {
			_ = fmt.Errorf("failed parse restore flag: %w", err)
		}
		config.Restore = boolValue
	}
	return config
}

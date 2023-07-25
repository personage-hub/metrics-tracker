package main

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
}

func parseFlags() Config {
	var config Config
	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Address of the HTTP server endpoint")

	if envValue := os.Getenv("ADDRESS"); envValue != "" {
		config.ServerAddress = envValue
	}
	flag.Parse()

	return config
}

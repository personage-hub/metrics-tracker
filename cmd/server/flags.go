package main

import (
	"flag"
	"os"
)

var serverAddress string

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "Address of the HTTP server endpoint")

	if envValue := os.Getenv("ADDRESS"); envValue != "" {
		serverAddress = envValue
	}

	flag.Parse()
}

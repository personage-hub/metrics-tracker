package main

import (
	"github.com/personage-hub/metrics-tracker/internal"
)

func main() {
	storage := internal.NewMemStorage()

	config := parseFlags()

	server := NewServer(storage)

	if err := server.Run(config); err != nil {
		panic(err)
	}
}

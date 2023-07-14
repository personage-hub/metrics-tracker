package main

import (
	"flag"
	"time"
)

var (
	serverAddress  string
	reportInterval time.Duration
	pollInterval   time.Duration
)

func init() {

	flag.StringVar(&serverAddress, "a", "localhost:8080", "Address of the HTTP server endpoint")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "Report interval for sending metrics to the server")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "Poll interval for collecting metrics")
	flag.Parse()
}

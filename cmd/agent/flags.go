package main

import (
	"flag"
	"time"
)

var (
	serverAddress       string
	reportIntervalParam int
	pollIntervalParam   int
	reportInterval      time.Duration
	pollInterval        time.Duration
)

func parseFlag() {

	flag.StringVar(&serverAddress, "a", "localhost:8080", "Address of the HTTP server endpoint")
	flag.IntVar(&reportIntervalParam, "r", 10, "Report interval for sending metrics to the server")
	flag.IntVar(&pollIntervalParam, "p", 2, "Poll interval for collecting metrics")
	flag.Parse()
	reportInterval = time.Duration(reportIntervalParam) * time.Second
	pollInterval = time.Duration(pollIntervalParam) * time.Second
}

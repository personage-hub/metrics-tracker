package main

import (
	"flag"
	"os"
	"strconv"
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

	if envValue := os.Getenv("ADDRESS"); envValue != "" {
		serverAddress = envValue
	}
	if envValue := os.Getenv("REPORT_INTERVAL"); envValue != "" {
		if intValue, err := strconv.Atoi(envValue); err == nil {
			reportIntervalParam = intValue
		}
	}
	if envValue := os.Getenv("POLL_INTERVAL"); envValue != "" {
		if intValue, err := strconv.Atoi(envValue); err == nil {
			pollIntervalParam = intValue
		}
	}

	reportInterval = time.Duration(reportIntervalParam) * time.Second
	pollInterval = time.Duration(pollIntervalParam) * time.Second
	reportInterval = time.Duration(reportIntervalParam) * time.Second
	pollInterval = time.Duration(pollIntervalParam) * time.Second
}

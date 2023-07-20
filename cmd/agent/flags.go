package main

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddress       string
	ReportIntervalParam int
	PollIntervalParam   int
	ReportInterval      time.Duration
	PollInterval        time.Duration
}

func parseFlag() Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Address of the HTTP server endpoint")
	flag.IntVar(&config.ReportIntervalParam, "r", 10, "Report interval for sending metrics to the server")
	flag.IntVar(&config.PollIntervalParam, "p", 2, "Poll interval for collecting metrics")
	flag.Parse()

	if envValue := os.Getenv("ADDRESS"); envValue != "" {
		config.ServerAddress = envValue
	}
	if envValue := os.Getenv("REPORT_INTERVAL"); envValue != "" {
		if intValue, err := strconv.Atoi(envValue); err == nil {
			config.ReportIntervalParam = intValue
		}
	}
	if envValue := os.Getenv("POLL_INTERVAL"); envValue != "" {
		if intValue, err := strconv.Atoi(envValue); err == nil {
			config.PollIntervalParam = intValue
		}
	}

	config.ReportInterval = time.Duration(config.ReportIntervalParam) * time.Second
	config.PollInterval = time.Duration(config.PollIntervalParam) * time.Second

	return config
}

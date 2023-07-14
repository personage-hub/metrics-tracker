package main

import (
	"flag"
)

var serverAddress string

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "Address of the HTTP server endpoint")
}

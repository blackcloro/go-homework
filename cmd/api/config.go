package main

import (
	"flag"
	"os"
)

// Ports can be specified with environment variables or command line flags. DEFAULT PORT 3000
func getPort() string {
	var port string
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	} else {
		port = ":3000"
	}

	flag.StringVar(&port, "port", port, "Port to listen on")
	flag.Parse()

	return port
}

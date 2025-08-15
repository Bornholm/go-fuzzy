package main

import "flag"

// Configuration for the server
type Config struct {
	Address     string
	Definitions string
}

func parseConfig() *Config {
	config := &Config{}

	// Parse command line flags
	flag.StringVar(&config.Address, "port", ":3003", "address to listen on")
	flag.StringVar(&config.Definitions, "definitions", "*.fuzzy", "dsl file pattern to load")
	flag.Parse()

	return config
}

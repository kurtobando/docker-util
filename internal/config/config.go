package config

import "flag"

// Config holds all application configuration
type Config struct {
	DBPath     string
	ServerPort string
}

// Load parses command line flags and returns configuration
// This preserves the exact same flags and defaults from the original main.go
func Load() *Config {
	config := &Config{}

	flag.StringVar(&config.DBPath, "dbpath", "./docker_logs.db", "Path to the SQLite database file.")
	flag.StringVar(&config.ServerPort, "port", "9123", "Port for the web UI.")

	flag.Parse()

	return config
}

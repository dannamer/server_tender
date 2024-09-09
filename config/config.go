package config

import (
	"os"
)

type Config struct {
	ServerAddress   string
	PostgresConn    string
	PostgresJDBCURL string
	PostgresUser    string
	PostgresPass    string
	PostgresHost    string
	PostgresPort    string
	PostgresDB      string
}

func LoadConfig() *Config {
	return &Config{
		ServerAddress:   os.Getenv("SERVER_ADDRESS"),
		PostgresConn:    os.Getenv("POSTGRES_CONN"),
		PostgresJDBCURL: os.Getenv("POSTGRES_JDBC_URL"),
		PostgresUser:    os.Getenv("POSTGRES_USERNAME"),
		PostgresPass:    os.Getenv("POSTGRES_PASSWORD"),
		PostgresHost:    os.Getenv("POSTGRES_HOST"),
		PostgresPort:    os.Getenv("POSTGRES_PORT"),
		PostgresDB:      os.Getenv("POSTGRES_DATABASE"),
	}
}

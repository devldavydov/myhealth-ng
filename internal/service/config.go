package service

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultHost         = "127.0.0.1"
	defaultPort         = 3000
	defaultClientOrigin = "http://localhost:5173"
)

type Config struct {
	Host                     string
	Port                     int
	ClientOrigin             string
	RequireClientCertificate bool
}

func ConfigFromEnvironment() (Config, error) {
	portValue := os.Getenv("PORT")
	if portValue == "" {
		portValue = strconv.Itoa(defaultPort)
	}
	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("некорректный PORT: %q", portValue)
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = defaultHost
	}
	clientOrigin := os.Getenv("CLIENT_ORIGIN")
	if clientOrigin == "" {
		clientOrigin = defaultClientOrigin
	}

	return Config{
		Host:                     host,
		Port:                     port,
		ClientOrigin:             clientOrigin,
		RequireClientCertificate: os.Getenv("REQUIRE_CLIENT_CERT") == "true",
	}, nil
}

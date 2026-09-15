package service

import (
	"flag"
	"fmt"
	"io"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = 3000
)

type Config struct {
	Host                     string
	Port                     int
	DatabaseURL              string
	RequireClientCertificate bool
}

func ConfigFromArgs(args []string) (Config, error) {
	config := Config{}
	flags := flag.NewFlagSet("myhealth-server", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&config.Host, "host", defaultHost, "адрес HTTP-сервера")
	flags.IntVar(&config.Port, "port", defaultPort, "порт HTTP-сервера")
	flags.StringVar(&config.DatabaseURL, "database-url", "", "строка подключения PostgreSQL")
	flags.BoolVar(&config.RequireClientCertificate, "require-client-cert", false, "требовать проверенный клиентский сертификат")
	if err := flags.Parse(args); err != nil {
		return Config{}, fmt.Errorf("параметры запуска: %w", err)
	}
	if flags.NArg() > 0 {
		return Config{}, fmt.Errorf("неожиданные аргументы: %v", flags.Args())
	}
	if config.Host == "" {
		return Config{}, fmt.Errorf("--host не может быть пустым")
	}
	if config.Port < 1 || config.Port > 65535 {
		return Config{}, fmt.Errorf("некорректный --port: %d", config.Port)
	}
	if config.DatabaseURL == "" {
		return Config{}, fmt.Errorf("обязателен --database-url")
	}
	return config, nil
}

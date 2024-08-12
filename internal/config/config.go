package config

import (
	"flag"
	"os"
)

type Flags struct {
	RunAddr          string
	RedirectBaseAddr string
	FileStoragePath  string
	DatabaseURI      string
}

func ParseConfig() Flags {
	flags := Flags{}
	flag.StringVar(&flags.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&flags.RedirectBaseAddr, "b", "http://localhost:8080/", "base for short links")
	flag.StringVar(&flags.FileStoragePath, "f", "index.json", "path to file storage")
	flag.StringVar(&flags.DatabaseURI, "d", "", "db connection string")
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDR"); envRunAddr != "" {
		flags.RunAddr = envRunAddr
	}
	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		flags.RedirectBaseAddr = envBaseAddr
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		flags.FileStoragePath = envFileStoragePath
	}
	if envDatabaseURI := os.Getenv("DATABASE_DSN"); envDatabaseURI != "" {
		flags.DatabaseURI = envDatabaseURI
	}

	return flags
}

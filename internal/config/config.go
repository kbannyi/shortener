package config

import (
	"flag"
	"os"
)

type Flags struct {
	RunAddr          string
	RedirectBaseAddr string
	FileStoragePath  string
}

func ParseConfig() Flags {
	flags := Flags{}
	flag.StringVar(&flags.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&flags.RedirectBaseAddr, "b", "http://localhost:8080/", "base for short links")
	flag.StringVar(&flags.FileStoragePath, "f", "index.json", "path to file storage")
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDR"); envRunAddr != "" {
		flags.RunAddr = envRunAddr
	}
	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		flags.RedirectBaseAddr = envBaseAddr
	}
	if envBaseAddr := os.Getenv("FILE_STORAGE_PATH"); envBaseAddr != "" {
		flags.FileStoragePath = envBaseAddr
	}

	return flags
}

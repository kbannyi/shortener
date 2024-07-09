package main

import (
	"fmt"
	"net/http"

	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/repository"
	"github.com/kbannyi/shortener/internal/router"
	"github.com/kbannyi/shortener/internal/service"
)

func main() {
	flags := config.ParseConfig()

	fmt.Println("Starting server...")
	err := http.ListenAndServe(flags.RunAddr,
		router.NewURLRouter(service.NewService(repository.NewRepository()), flags))
	if err != nil {
		panic(err)
	}
}

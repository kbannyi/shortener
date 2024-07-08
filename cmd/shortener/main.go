package main

import (
	"fmt"
	"net/http"

	"github.com/kbannyi/shortener/internal/repository"
	"github.com/kbannyi/shortener/internal/router"
	"github.com/kbannyi/shortener/internal/service"
)

func main() {
	fmt.Println("Starting server...")
	err := http.ListenAndServe(":8080",
		router.NewURLRouter(service.NewService(repository.NewRepository())))
	if err != nil {
		panic(err)
	}
}

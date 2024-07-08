package main

import (
	"fmt"
	"net/http"

	"github.com/kbannyi/shortener/internal/handler"
	"github.com/kbannyi/shortener/internal/repository"
	"github.com/kbannyi/shortener/internal/service"
)

func main() {
	mux := http.NewServeMux()
	h := handler.NewHandler(service.NewService(repository.NewRepository()))
	mux.Handle("/", h)

	fmt.Println("Starting server...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

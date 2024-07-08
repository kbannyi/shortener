package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/kbannyi/shortener/internal/handler"
	"github.com/kbannyi/shortener/internal/repository"
	"github.com/kbannyi/shortener/internal/service"
)

func main() {
	fmt.Println("Starting server...")
	err := http.ListenAndServe(":8080", URLRouter())
	if err != nil {
		panic(err)
	}
}

func URLRouter() chi.Router {
	r := chi.NewRouter()
	h := handler.NewHandler(service.NewService(repository.NewRepository()))

	r.Get("/{id}", h.HandleGet)
	r.Post("/", h.HandlePost)

	return r
}

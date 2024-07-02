package main

import (
	"fmt"
	"net/http"

	"github.com/kbannyi/shortener/internal/handler/url"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", url.NewHandler())

	fmt.Println("Starting server...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/logger"
	"github.com/kbannyi/shortener/internal/middleware"
	"github.com/kbannyi/shortener/internal/repository"
	"github.com/kbannyi/shortener/internal/router"
	"github.com/kbannyi/shortener/internal/service"
)

func main() {
	flags := config.ParseConfig()

	if err := logger.Initialize("Debug"); err != nil {
		fmt.Printf("Coudn't initialize logger: %v\n", err)
		return
	}

	repo, err := repository.NewRepository(flags)
	if err != nil {
		logger.Log.Errorf("Coudn't initialize repository: %v", err)
		return
	}
	serv := service.NewService(repo)
	r := chi.NewRouter()
	r.Use(middleware.RequestLoggerMiddleware)
	r.Use(middleware.ResponseLoggerMiddleware)
	r.Use(middleware.GZIPMiddleware)
	r.Mount("/", router.NewURLRouter(serv, flags))

	logger.Log.Info("Starting server...")
	logger.Log.Infow("Running on:", "url", flags.RunAddr)
	logger.Log.Infow("Base for short links:", "url", flags.RedirectBaseAddr)
	logger.Log.Infow("Using storage file:", "path", flags.FileStoragePath)
	if http.ListenAndServe(flags.RunAddr, r) != nil {
		logger.Log.Errorf("Error on serve: %v", err)
	}
}

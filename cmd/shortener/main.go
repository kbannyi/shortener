package main

import (
	"net/http"

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
		logger.Log.Errorf("Coudn't initialize logger: %v", err)
	}
	logger.Log.Infow("Running on:", "url", flags.RunAddr)
	logger.Log.Infow("Base for short links:", "url", flags.RedirectBaseAddr)

	logger.Log.Info("Starting server...")
	var h http.Handler = router.NewURLRouter(service.NewService(repository.NewRepository()), flags)
	h = middleware.ResponseLoggerMiddleware(h)
	h = middleware.RequestLoggerMiddleware(h)
	h = middleware.GZIPMiddleware(h)
	err := http.ListenAndServe(flags.RunAddr, h)
	if err != nil {
		logger.Log.Error("Error on serve: %v", err)
	}
}

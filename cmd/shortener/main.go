package main

import (
	"net/http"

	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/logger"
	"github.com/kbannyi/shortener/internal/repository"
	"github.com/kbannyi/shortener/internal/router"
	"github.com/kbannyi/shortener/internal/service"
)

func main() {
	flags := config.ParseConfig()
	if err := logger.Initialize("Debug"); err != nil {
		panic(err)
	}
	logger.Log.Info("Running on:", "url", flags.RunAddr)
	logger.Log.Info("Base for short links:", "url", flags.RedirectBaseAddr)

	logger.Log.Info("Starting server...")
	var h http.Handler = router.NewURLRouter(service.NewService(repository.NewRepository()), flags)
	h = logger.ResponseLogger(h)
	h = logger.RequestLogger(h)
	err := http.ListenAndServe(flags.RunAddr, h)
	if err != nil {
		panic(err)
	}
}

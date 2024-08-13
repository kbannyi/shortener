package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/dbtools"
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

	var db *sql.DB
	var dbx *sqlx.DB
	var err error
	if flags.DatabaseURI != "" {
		db, err = sql.Open("pgx", flags.DatabaseURI)
		if err != nil {
			logger.Log.Errorf("Unable to connect to database: %v\n", err)
			return
		}
		defer db.Close()
		err = dbtools.MigrateDB(db)
		if err != nil {
			logger.Log.Errorf("Unable to apply database migrations: %v\n", err)
			return
		}
		dbx = sqlx.NewDb(db, "pgx")
	}
	var repo service.Repository
	switch {
	case dbx != nil:
		repo, err = repository.NewPostgresUserRepository(dbx)
		logger.Log.Info("Using DB storage")
	case flags.FileStoragePath != "":
		repo, err = repository.NewFileURLRepository(flags)
		logger.Log.Infow("Using storage file:", "path", flags.FileStoragePath)
	default:
		repo, err = repository.NewMemoryURLRepository()
		logger.Log.Info("Using InMemory storage")
	}
	if err != nil {
		logger.Log.Errorf("Coudn't initialize repository: %v", err)
		return
	}
	serv := service.NewService(repo)
	r := chi.NewRouter()
	r.Use(middleware.RequestLoggerMiddleware)
	r.Use(middleware.ResponseLoggerMiddleware)
	r.Use(middleware.GZIPMiddleware)
	r.Get("/ping", ping(db))
	r.Mount("/", router.NewURLRouter(serv, flags))

	logger.Log.Info("Starting server...")
	logger.Log.Infow("Running on:", "url", flags.RunAddr)
	logger.Log.Infow("Base for short links:", "url", flags.RedirectBaseAddr)
	if http.ListenAndServe(flags.RunAddr, r) != nil {
		logger.Log.Errorf("Error on serve: %v", err)
	}
}

func ping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "failed to connect to db", http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "failed to connect to db", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

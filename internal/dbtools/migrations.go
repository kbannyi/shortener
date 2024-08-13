package dbtools

import (
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kbannyi/shortener/internal/logger"
)

func MigrateDB(db *sql.DB) error {
	migrationsFolder := "db/migrations"
	files, err := os.ReadDir(migrationsFolder)
	if err != nil {
		return err
	}
	logger.Log.Debug("migrating db...")
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, file := range files {
		err = applysql(tx, filepath.Join(migrationsFolder, file.Name()))
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	logger.Log.Debug("migrated succesfully")

	return nil
}

func applysql(tx *sql.Tx, name string) error {
	logger.Log.Debugf("applying %q...", name)
	sql, err := readsql(name)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancel()
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return err
	}

	return nil
}

func readsql(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	sqlbytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(sqlbytes), nil
}

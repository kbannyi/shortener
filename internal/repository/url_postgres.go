package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/kbannyi/shortener/internal/domain"
)

type PostgresURLRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) (*PostgresURLRepository, error) {
	repo := &PostgresURLRepository{db}

	return repo, nil
}

func (r *PostgresURLRepository) Save(ctx context.Context, url *domain.URL) error {
	_, err := r.db.NamedExecContext(ctx, `INSERT INTO url (id, short_url, original_url)
	VALUES (:id, :short_url, :original_url)`, url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {

			return &DuplicateURLError{URL: url}
		}
	}

	return nil
}

func (r *PostgresURLRepository) BatchSave(ctx context.Context, urls []*domain.URL) error {
	_, err := r.db.NamedExecContext(ctx, `INSERT INTO url (id, short_url, original_url)
	VALUES (:id, :short_url, :original_url)`, urls)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresURLRepository) Get(ctx context.Context, id string) (*domain.URL, bool) {
	URL := domain.URL{}
	err := r.db.GetContext(ctx, &URL, `SELECT * FROM url WHERE id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, false
	}

	return &URL, true
}

package repository

import (
	"context"
	"database/sql"
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
	_, err := r.db.NamedExecContext(ctx, `INSERT INTO url (id, short_url, original_url, user_id)
	VALUES (:id, :short_url, :original_url, :user_id)`, url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return &ErrDuplicateURL{URL: url}
		}
		return err
	}

	return nil
}

func (r *PostgresURLRepository) BatchSave(ctx context.Context, urls []*domain.URL) error {
	_, err := r.db.NamedExecContext(ctx, `INSERT INTO url (id, short_url, original_url, user_id)
	VALUES (:id, :short_url, :original_url, :user_id)`, urls)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresURLRepository) Get(ctx context.Context, id string) (*domain.URL, error) {
	url := domain.URL{}
	err := r.db.GetContext(ctx, &url, `SELECT * FROM url WHERE id = $1 LIMIT 1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if url.IsDeleted {
		return nil, ErrDeleted
	}

	return &url, nil
}

func (r *PostgresURLRepository) GetByUser(ctx context.Context, userid string) ([]*domain.URL, error) {
	urls := []*domain.URL{}
	err := r.db.SelectContext(ctx, &urls, "SELECT * FROM url WHERE user_id=$1", userid)

	return urls, err
}

func (r *PostgresURLRepository) GetList(ctx context.Context, ids []string) ([]*domain.URL, error) {
	query, args, err := sqlx.In("SELECT * FROM url WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	urls := make([]*domain.URL, 0, len(ids))
	err = r.db.SelectContext(ctx, &urls, query, args...)
	if err != nil {
		return nil, err
	}
	if len(urls) != len(ids) {
		return nil, ErrNotFound
	}

	return urls, nil
}

func (r *PostgresURLRepository) DeleteIDs(ctx context.Context, ids []string) error {
	query, args, err := sqlx.In("UPDATE url SET is_deleted= TRUE WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

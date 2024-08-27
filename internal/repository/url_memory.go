package repository

import (
	"context"
	"sync"

	"github.com/kbannyi/shortener/internal/domain"
)

type MemoryURLRepository struct {
	byID map[string]*domain.URL
	mu   sync.RWMutex
}

func NewMemoryURLRepository() (*MemoryURLRepository, error) {
	return &MemoryURLRepository{
		byID: make(map[string]*domain.URL),
	}, nil
}

func (r *MemoryURLRepository) Save(ctx context.Context, url *domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.byID[url.ID]
	if ok {
		return &ErrDuplicateURL{URL: url}
	}
	r.byID[url.ID] = url

	return nil
}

func (r *MemoryURLRepository) BatchSave(ctx context.Context, urls []*domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, url := range urls {
		_, ok := r.byID[url.ID]
		if ok {
			continue
		}

		r.byID[url.ID] = url
	}

	return nil
}

func (r *MemoryURLRepository) Get(ctx context.Context, ID string) (*domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.byID[ID]
	if !ok {
		return nil, ErrNotFound
	}
	if url.IsDeleted {
		return nil, ErrDeleted
	}

	return url, nil
}

func (r *MemoryURLRepository) GetByUser(ctx context.Context, userid string) ([]*domain.URL, error) {
	urls := []*domain.URL{}
	for _, url := range r.byID {
		if url.UserID != nil && *url.UserID == userid {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

func (r *MemoryURLRepository) GetList(ctx context.Context, ids []string) ([]*domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*domain.URL, 0, len(ids))
	for _, id := range ids {
		url, ok := r.byID[id]
		if !ok {
			return nil, ErrNotFound
		}
		if url.IsDeleted {
			return nil, ErrDeleted
		}
		results = append(results, url)
	}

	return results, nil
}

func (r *MemoryURLRepository) DeleteIDs(ctx context.Context, ids []string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*domain.URL, 0, len(ids))
	for _, id := range ids {
		url, ok := r.byID[id]
		if !ok {
			return ErrNotFound
		}
		if url.IsDeleted {
			return ErrDeleted
		}
		results = append(results, url)
	}

	for _, url := range results {
		url.IsDeleted = true
	}

	return nil
}

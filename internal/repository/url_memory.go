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
		return &DuplicateURLError{URL: url}
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

func (r *MemoryURLRepository) Get(ctx context.Context, ID string) (URL *domain.URL, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	URL, ok = r.byID[ID]

	return
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

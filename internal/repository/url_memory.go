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

func (r *MemoryURLRepository) Save(ctx context.Context, URL *domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[URL.ID] = URL

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

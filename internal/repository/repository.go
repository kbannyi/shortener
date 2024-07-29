package repository

import (
	"sync"

	"github.com/kbannyi/shortener/internal/domain"
)

type URLRepository struct {
	byID map[string]*domain.URL
	mu   sync.RWMutex
}

func NewRepository() *URLRepository {
	return &URLRepository{
		byID: make(map[string]*domain.URL),
	}
}

func (r *URLRepository) Save(URL *domain.URL) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[URL.ID] = URL
}

func (r *URLRepository) Get(ID string) (URL *domain.URL, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	URL, ok = r.byID[ID]

	return
}

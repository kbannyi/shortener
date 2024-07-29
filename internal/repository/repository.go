package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/domain"
)

type URLRepository struct {
	byID map[string]*domain.URL
	mu   sync.RWMutex
}

func NewRepository(flags config.Flags) (*URLRepository, error) {
	repo := &URLRepository{
		byID: make(map[string]*domain.URL),
	}

	if err := repo.readIndex(flags.FileStoragePath); err != nil {
		return nil, fmt.Errorf("couldn't open storage file: %w", err)
	}

	return repo, nil
}

func (r *URLRepository) readIndex(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		data := scanner.Bytes()
		url := domain.URL{}
		if err := json.Unmarshal(data, &url); err != nil {
			return err
		}
		r.byID[url.ID] = &url
	}

	return nil
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

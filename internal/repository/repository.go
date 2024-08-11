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
	byID            map[string]*domain.URL
	mu              sync.RWMutex
	fileStoragePath string
}

func NewRepository(flags config.Flags) (*URLRepository, error) {
	repo := &URLRepository{
		byID:            make(map[string]*domain.URL),
		fileStoragePath: flags.FileStoragePath,
	}

	if err := repo.readIndex(); err != nil {
		return nil, fmt.Errorf("couldn't open storage file: %w", err)
	}

	return repo, nil
}

func (r *URLRepository) readIndex() error {
	f, err := os.OpenFile(r.fileStoragePath, os.O_CREATE|os.O_RDWR, 0o666)
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

func (r *URLRepository) Save(url *domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.byID[url.ID]
	if ok {
		return nil
	}

	if err := saveToIndex(url, r.fileStoragePath); err != nil {
		return fmt.Errorf("couldn't write to storage file: %w", err)
	}
	r.byID[url.ID] = url

	return nil
}

func saveToIndex(url *domain.URL, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	defer file.Close()
	var data []byte
	data, err = json.Marshal(url)
	if err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		return err
	}
	if err := writer.WriteByte('\n'); err != nil {
		return err
	}

	return writer.Flush()
}

func (r *URLRepository) Get(ID string) (URL *domain.URL, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	URL, ok = r.byID[ID]

	return
}

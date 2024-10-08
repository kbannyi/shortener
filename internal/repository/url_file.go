package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/domain"
)

type FileURLRepository struct {
	byID            map[string]*domain.URL
	mu              sync.RWMutex
	fileStoragePath string
}

func NewFileURLRepository(flags config.Flags) (*FileURLRepository, error) {
	repo := &FileURLRepository{
		byID:            make(map[string]*domain.URL),
		fileStoragePath: flags.FileStoragePath,
	}

	if err := repo.readIndex(); err != nil {
		return nil, fmt.Errorf("couldn't open storage file: %w", err)
	}

	return repo, nil
}

func (r *FileURLRepository) readIndex() error {
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

func (r *FileURLRepository) Save(ctx context.Context, url *domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.byID[url.ID]
	if ok {
		return &DuplicateURLError{URL: url}
	}

	if err := saveToIndex([]*domain.URL{url}, r.fileStoragePath); err != nil {
		return fmt.Errorf("couldn't write to storage file: %w", err)
	}
	r.byID[url.ID] = url

	return nil
}

func (r *FileURLRepository) BatchSave(ctx context.Context, urls []*domain.URL) error {
	batch := make([]*domain.URL, 0, len(urls))
	for _, url := range urls {
		_, ok := r.byID[url.ID]
		if ok {
			continue
		}
		batch = append(batch, url)
	}

	if len(batch) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := saveToIndex(batch, r.fileStoragePath); err != nil {
		return fmt.Errorf("couldn't write to storage file: %w", err)
	}

	for _, url := range batch {
		r.byID[url.ID] = url
	}

	return nil
}

func saveToIndex(urls []*domain.URL, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	defer file.Close()
	for _, url := range urls {
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
	}

	return writer.Flush()
}

func (r *FileURLRepository) Get(ctx context.Context, ID string) (URL *domain.URL, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	URL, ok = r.byID[ID]

	return
}

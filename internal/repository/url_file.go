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
		return &ErrDuplicateURL{URL: url}
	}

	if err := saveToIndex([]*domain.URL{url}, r.fileStoragePath); err != nil {
		return fmt.Errorf("couldn't write to storage file: %w", err)
	}
	r.byID[url.ID] = url

	return nil
}

func (r *FileURLRepository) BatchSave(ctx context.Context, urls []*domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := saveToIndex(urls, r.fileStoragePath); err != nil {
		return fmt.Errorf("couldn't write to storage file: %w", err)
	}

	for _, url := range urls {
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

func (r *FileURLRepository) Get(ctx context.Context, ID string) (*domain.URL, error) {
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

func (r *FileURLRepository) GetByUser(ctx context.Context, userid string) ([]*domain.URL, error) {
	urls := []*domain.URL{}
	for _, url := range r.byID {
		if url.UserID != nil && *url.UserID == userid {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

func (r *FileURLRepository) GetList(ctx context.Context, ids []string) ([]*domain.URL, error) {
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

func (r *FileURLRepository) DeleteIDs(ctx context.Context, ids []string) error {
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
	err := r.BatchSave(ctx, results)
	if err != nil {
		return err
	}

	return nil
}

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/kbannyi/shortener/internal/domain"
	"github.com/kbannyi/shortener/internal/models"
)

type Repository interface {
	Save(ctx context.Context, url *domain.URL) error
	BatchSave(ctx context.Context, urls []*domain.URL) error
	Get(ctx context.Context, id string) (*domain.URL, bool)
}

type URLService struct {
	Repository Repository
}

func NewService(r Repository) *URLService {
	return &URLService{r}
}

func (s *URLService) Create(value string) (ID string, err error) {
	URL := domain.NewURL(value)
	if err := s.Repository.Save(context.TODO(), URL); err != nil {
		return "", fmt.Errorf("coudln't save domain.URL: %w", err)
	}

	return URL.ID, nil
}

func (s *URLService) Get(ID string) (string, bool) {
	v, ok := s.Repository.Get(context.TODO(), ID)
	if !ok {
		return "", ok
	}

	return v.Original, ok
}

func (s *URLService) BatchCreate(ctx context.Context, correlated []models.CorrelatedURL) (map[string]*domain.URL, error) {
	results := make(map[string]*domain.URL, len(correlated))
	batch := make([]*domain.URL, 0, len(correlated))
	for _, orig := range correlated {
		_, ok := results[orig.CorrelationID]
		if ok {
			return nil, errors.New("correlationId duplicate")
		}
		url := domain.NewURL(orig.Value)
		results[orig.CorrelationID] = url
		batch = append(batch, url)
	}
	err := s.Repository.BatchSave(ctx, batch)
	if err != nil {
		return nil, err
	}

	return results, nil
}

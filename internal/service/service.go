package service

import (
	"context"
	"fmt"

	"github.com/kbannyi/shortener/internal/domain"
)

type Repository interface {
	Save(ctx context.Context, url *domain.URL) error
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

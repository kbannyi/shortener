package service

import (
	"github.com/kbannyi/shortener/internal/domain"
)

type Repository interface {
	Save(*domain.URL)
	Get(ID string) (*domain.URL, bool)
}

type URLService struct {
	Repository Repository
}

func NewService(r Repository) *URLService {
	return &URLService{r}
}

func (s *URLService) Create(value string) (ID string) {
	URL := domain.NewURL(value)
	s.Repository.Save(URL)

	return URL.ID
}

func (s *URLService) Get(ID string) (string, bool) {
	v, ok := s.Repository.Get(ID)
	if !ok {
		return "", ok
	}

	return v.Value, ok
}

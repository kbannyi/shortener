package url

import (
	"github.com/kbannyi/shortener/internal/domain"
	"github.com/kbannyi/shortener/internal/repository/url"
)

type URLService struct {
	repository url.URLRepository
}

func NewService() *URLService {
	return &URLService{
		repository: *url.NewRepository(),
	}
}

func (s *URLService) Create(value string) (ID string) {
	URL := domain.NewURL(value)
	s.repository.Save(URL)

	return URL.ID
}

func (s *URLService) Get(ID string) (string, bool) {
	v, ok := s.repository.Get(ID)
	if !ok {
		return "", ok
	}

	return v.Value, ok
}

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kbannyi/shortener/internal/auth"
	"github.com/kbannyi/shortener/internal/domain"
	"github.com/stretchr/testify/assert"
)

var TestURL = &domain.URL{
	ID:       "testid",
	Original: "linkvalue",
}

type MockRepository struct{}

func (r *MockRepository) GetByUser(ctx context.Context, id string) ([]*domain.URL, error) {
	panic("unimplemented")
}
func (r *MockRepository) Save(ctx context.Context, URL *domain.URL) error { return nil }

func (r *MockRepository) Get(ctx context.Context, ID string) (URL *domain.URL, ok bool) {
	URL = TestURL
	ok = ID == URL.ID

	return
}

func (r *MockRepository) BatchSave(ctx context.Context, urls []*domain.URL) error {
	return errors.New("not implemented")
}

func TestCreate_ReturnsNonEmptyId(t *testing.T) {
	s := NewService(&MockRepository{})

	ID, err := s.Create(auth.ToContext(context.Background(), auth.AuthUser{}), TestURL.Original)
	assert.NoError(t, err)
	assert.NotEmpty(t, ID)
}

func TestGet_ReturnsKnownValue(t *testing.T) {
	s := NewService(&MockRepository{})
	v1, ok1 := s.Get(TestURL.ID)
	v2, ok2 := s.Get("unknownid")

	assert.Equal(t, v1, TestURL.Original)
	assert.True(t, ok1)
	assert.Empty(t, v2)
	assert.False(t, ok2)
}

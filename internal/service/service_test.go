package service

import (
	"context"
	"testing"

	"github.com/kbannyi/shortener/internal/domain"
	"github.com/stretchr/testify/assert"
)

type MockRepository struct{}

var TestURL = &domain.URL{
	ID:       "testid",
	Original: "linkvalue",
}

func (r *MockRepository) Save(ctx context.Context, URL *domain.URL) error { return nil }

func (r *MockRepository) Get(ctx context.Context, ID string) (URL *domain.URL, ok bool) {
	URL = TestURL
	ok = ID == URL.ID

	return
}

func TestCreate_ReturnsNonEmptyId(t *testing.T) {
	s := NewService(&MockRepository{})

	ID, err := s.Create(TestURL.Original)
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

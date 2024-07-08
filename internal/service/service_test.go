package service

import (
	"testing"

	"github.com/kbannyi/shortener/internal/domain"
	"github.com/stretchr/testify/assert"
)

type MockRepository struct{}

var TestURL = &domain.URL{
	ID:    "testid",
	Value: "linkvalue",
}

func (r *MockRepository) Save(URL *domain.URL) {}

func (r *MockRepository) Get(ID string) (URL *domain.URL, ok bool) {
	URL = TestURL
	ok = ID == URL.ID

	return
}

func TestCreate_ReturnsNonEmptyId(t *testing.T) {
	s := NewService(&MockRepository{})

	ID := s.Create(TestURL.Value)
	assert.NotEmpty(t, ID)
}

func TestGet_ReturnsKnownValue(t *testing.T) {
	s := NewService(&MockRepository{})
	v1, ok1 := s.Get(TestURL.ID)
	v2, ok2 := s.Get("unknownid")

	assert.Equal(t, v1, TestURL.Value)
	assert.True(t, ok1)
	assert.Empty(t, v2)
	assert.False(t, ok2)
}

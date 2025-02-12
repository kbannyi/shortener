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

func (r *MockRepository) DeleteIDs(ctx context.Context, ids []string) error {
	panic("unimplemented")
}

func (r *MockRepository) GetList(ctx context.Context, ids []string) ([]*domain.URL, error) {
	panic("unimplemented")
}

func (r *MockRepository) GetByUser(ctx context.Context, id string) ([]*domain.URL, error) {
	panic("unimplemented")
}
func (r *MockRepository) Save(ctx context.Context, URL *domain.URL) error { return nil }

func (r *MockRepository) Get(ctx context.Context, ID string) (URL *domain.URL, err error) {
	URL = TestURL
	err = nil
	if ID != URL.ID {
		err = errors.New("not found")
	}

	return
}

func (r *MockRepository) BatchSave(ctx context.Context, urls []*domain.URL) error {
	return errors.New("not implemented")
}

func TestCreate_ReturnsNonEmptyId(t *testing.T) {
	s := NewService(context.Background(), &MockRepository{})

	ID, err := s.Create(auth.ToContext(context.Background(), auth.AuthUser{}), TestURL.Original)
	assert.NoError(t, err)
	assert.NotEmpty(t, ID)
}

func TestGet_ReturnsKnownValue(t *testing.T) {
	s := NewService(context.Background(), &MockRepository{})
	v1, err1 := s.Get(TestURL.ID)
	v2, err2 := s.Get("unknownid")

	assert.Equal(t, v1, TestURL.Original)
	assert.NoError(t, err1)
	assert.Empty(t, v2)
	assert.Error(t, err2)
}

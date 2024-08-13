package repository

import (
	"context"
	"testing"

	"github.com/kbannyi/shortener/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMemoryURLRepository_ReturnsSavedURLs(t *testing.T) {
	r, err := NewMemoryURLRepository()
	assert.NoError(t, err)
	const testValue = "testvalue"

	URL := domain.NewURL(testValue)
	ctx := context.Background()
	err = r.Save(ctx, URL)
	assert.NoError(t, err)
	v1, ok1 := r.Get(ctx, URL.ID)
	v2, ok2 := r.Get(ctx, "unknownid")

	assert.Equal(t, v1.Original, testValue)
	assert.Equal(t, v1.ID, URL.ID)
	assert.True(t, ok1)
	assert.Nil(t, v2)
	assert.False(t, ok2)
}

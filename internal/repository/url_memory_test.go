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

func TestMemoryURLRepository_ReturnsBatchSavedURLs(t *testing.T) {
	r, err := NewMemoryURLRepository()
	assert.NoError(t, err)
	const testValue1 = "testvalue1"
	const testValue2 = "testvalue2"

	URL1 := domain.NewURL(testValue1)
	URL2 := domain.NewURL(testValue2)
	ctx := context.Background()
	err = r.BatchSave(ctx, []*domain.URL{URL1, URL2})
	assert.NoError(t, err)
	v1, ok1 := r.Get(ctx, URL1.ID)
	v2, ok2 := r.Get(ctx, URL2.ID)
	v3, ok3 := r.Get(ctx, "unknownid")

	assert.Equal(t, v1.Original, testValue1)
	assert.Equal(t, v1.ID, URL1.ID)
	assert.Equal(t, v2.Original, testValue2)
	assert.Equal(t, v2.ID, URL2.ID)
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.Nil(t, v3)
	assert.False(t, ok3)
}

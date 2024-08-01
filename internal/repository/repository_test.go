package repository

import (
	"testing"

	"github.com/kbannyi/shortener/internal/config"
	"github.com/kbannyi/shortener/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestRepository_ReturnsSavedURLs(t *testing.T) {
	tempdir := t.TempDir()
	r, err := NewRepository(config.Flags{FileStoragePath: tempdir + "/index_test.json"})
	assert.NoError(t, err)
	const testValue = "testvalue"

	URL := domain.NewURL(testValue)
	err = r.Save(URL)
	assert.NoError(t, err)
	v1, ok1 := r.Get(URL.ID)
	v2, ok2 := r.Get("unknownid")

	assert.Equal(t, v1.Original, testValue)
	assert.Equal(t, v1.ID, URL.ID)
	assert.True(t, ok1)
	assert.Nil(t, v2)
	assert.False(t, ok2)
}

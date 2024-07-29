package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewURL_SameValueSameID(t *testing.T) {
	url1 := NewURL("value")
	url2 := NewURL("value")
	url3 := NewURL("othervalue")

	assert.Equal(t, url1.ID, url2.ID)
	assert.NotEqual(t, url1.ID, url3.ID)
}

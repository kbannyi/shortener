package repository

import (
	"fmt"

	"github.com/kbannyi/shortener/internal/domain"
)

type DuplicateURLError struct {
	URL *domain.URL
	Err error
}

func (de *DuplicateURLError) Error() string {
	return fmt.Sprintf("attempt to add existing URL %q", de.URL.Original)
}

func (de *DuplicateURLError) Unwrap() error {
	return de.Err
}

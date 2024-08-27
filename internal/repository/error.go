package repository

import (
	"errors"
	"fmt"

	"github.com/kbannyi/shortener/internal/domain"
)

type ErrDuplicateURL struct {
	URL *domain.URL
	Err error
}

var ErrNotFound = errors.New("values not found")
var ErrDeleted = errors.New("requested value deleted")

func (de *ErrDuplicateURL) Error() string {
	return fmt.Sprintf("attempt to add existing URL %q", de.URL.Original)
}

func (de *ErrDuplicateURL) Unwrap() error {
	return de.Err
}

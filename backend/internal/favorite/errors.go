package favorite

import "errors"

var (
    ErrInvalidInput     = errors.New("invalid input")
    ErrFavoriteNotFound = errors.New("favorite not found")
    ErrAlreadyFavorited = errors.New("favorite already exists")
)


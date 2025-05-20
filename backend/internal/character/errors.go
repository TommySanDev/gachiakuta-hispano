package character

import "errors"

var (
    ErrCharacterNotFound = errors.New("character not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrCharacterAlreadyExists = errors.New("character already exists")
)

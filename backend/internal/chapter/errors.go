package chapter

import "errors"

var (
    ErrChapterNotFound     = errors.New("chapter not found")
    ErrInvalidInput        = errors.New("invalid input")
    ErrChapterAlreadyExists = errors.New("chapter already exists")
)


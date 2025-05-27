package comment

import "errors"

var (
    ErrInvalidInput     = errors.New("invalid input")
    ErrCommentNotFound  = errors.New("comment not found")
    ErrForbidden        = errors.New("action not allowed")
)


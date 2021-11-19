package db

import "errors"

var (
	ErrUnknown          = errors.New("unknown")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrInternal         = errors.New("internal error")
)

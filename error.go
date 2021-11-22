package dbw

import "errors"

var (
	ErrUnknown          = errors.New("unknown")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrInternal         = errors.New("internal error")
	ErrRecordNotFound   = errors.New("record not found")
)

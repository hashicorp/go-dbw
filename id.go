package db

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/go-secure-stdlib/base62"
	"golang.org/x/crypto/blake2b"
)

func NewPrivateId(prefix string, opt ...Option) (string, error) {
	return newId(prefix, opt...)
}

// NewPublicId creates a new public id with the prefix
func NewPublicId(prefix string, opt ...Option) (string, error) {
	return newId(prefix, opt...)
}

func newId(prefix string, opt ...Option) (string, error) {
	const op = "db.newId"
	if prefix == "" {
		return "", fmt.Errorf("%s: missing prefix: %w", op, ErrInvalidParameter)
	}
	var publicId string
	var err error
	opts := GetOpts(opt...)
	if len(opts.withPrngValues) > 0 {
		sum := blake2b.Sum256([]byte(strings.Join(opts.withPrngValues, "|")))
		reader := bytes.NewReader(sum[0:])
		publicId, err = base62.RandomWithReader(10, reader)
	} else {
		publicId, err = base62.Random(10)
	}
	if err != nil {
		return "", fmt.Errorf("%s: unable to generate id: %w", op, ErrInternal)
	}
	return fmt.Sprintf("%s_%s", prefix, publicId), nil
}

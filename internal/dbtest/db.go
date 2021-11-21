// Package db_test provides some helper funcs for testing db integrations
package dbtest

import (
	"github.com/hashicorp/go-secure-stdlib/base62"
	"google.golang.org/protobuf/proto"
)

const (
	defaultUserTablename = "db_test_user"
)

type TestUser struct {
	*StoreTestUser
	table string `gorm:"-"`
}

func NewTestUser() (*TestUser, error) {
	publicId, err := base62.Random(20)
	if err != nil {
		return nil, err
	}
	return &TestUser{
		StoreTestUser: &StoreTestUser{
			PublicId: publicId,
		},
	}, nil
}

func AllocTestUser() TestUser {
	return TestUser{
		StoreTestUser: &StoreTestUser{},
	}
}

// Clone is useful when you're retrying transactions and you need to send the user several times
func (u *TestUser) Clone() interface{} {
	s := proto.Clone(u.StoreTestUser)
	return &TestUser{
		StoreTestUser: s.(*StoreTestUser),
	}
}

func (u *TestUser) TableName() string {
	if u.table != "" {
		return u.table
	}
	return defaultUserTablename
}

func (u *TestUser) SetTableName(name string) {
	switch name {
	case "":
		u.table = defaultUserTablename
	default:
		u.table = name
	}
}

type Cloner interface {
	Clone() interface{}
}

type NotIder struct{}

func (i *NotIder) Clone() interface{} {
	return &NotIder{}
}

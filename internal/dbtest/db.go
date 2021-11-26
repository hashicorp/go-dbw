// Package db_test provides some helper funcs for testing db integrations
package dbtest

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-dbw"

	"github.com/hashicorp/go-secure-stdlib/base62"
	"google.golang.org/protobuf/proto"
)

const (
	defaultUserTablename   = "db_test_user"
	defaultCarTableName    = "db_test_car"
	defaultRentalTableName = "db_test_rental"
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

var _ dbw.VetForWriter = (*TestUser)(nil)

func (u *TestUser) VetForWrite(ctx context.Context, r dbw.Reader, opType dbw.OpType, opt ...dbw.Option) error {
	const op = "dbtest.(TestUser).VetForWrite"
	if u.PublicId == "" {
		return fmt.Errorf("%s: missing public id: %w", op, dbw.ErrInvalidParameter)
	}
	if u.Name == "fail-VetForWrite" {
		return fmt.Errorf("%s: name was fail-VetForWrite: %w", op, dbw.ErrInvalidParameter)
	}
	switch opType {
	case dbw.UpdateOp:
		dbOptions := dbw.GetOpts(opt...)
		for _, path := range dbOptions.WithFieldMaskPaths {
			switch path {
			case "PublicId", "CreateTime", "UpdateTime":
				return fmt.Errorf("%s: %s is immutable: %w", op, path, dbw.ErrInvalidParameter)
			}
		}
	case dbw.CreateOp:
		if u.CreateTime != nil {
			return fmt.Errorf("%s: create time is set by the database: %w", op, dbw.ErrInvalidParameter)
		}
	}
	return nil
}

type TestCar struct {
	*StoreTestCar
	table string `gorm:"-"`
}

func NewTestCar() (*TestCar, error) {
	publicId, err := base62.Random(20)
	if err != nil {
		return nil, err
	}
	return &TestCar{
		StoreTestCar: &StoreTestCar{
			PublicId: publicId,
		},
	}, nil
}

func (c *TestCar) TableName() string {
	if c.table != "" {
		return c.table
	}

	return defaultCarTableName
}

func (c *TestCar) SetTableName(name string) {
	c.table = name
}

type Cloner interface {
	Clone() interface{}
}

type NotIder struct{}

func (i *NotIder) Clone() interface{} {
	return &NotIder{}
}

type TestRental struct {
	*StoreTestRental
	table string `gorm:"-"`
}

func NewTestRental(userId, carId string) (*TestRental, error) {
	return &TestRental{
		StoreTestRental: &StoreTestRental{
			UserId: userId,
			CarId:  carId,
		},
	}, nil
}

func (r *TestRental) TableName() string {
	if r.table != "" {
		return r.table
	}

	return defaultRentalTableName
}

func (r *TestRental) SetTableName(name string) {
	r.table = name
}

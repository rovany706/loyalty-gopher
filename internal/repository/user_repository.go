package repository

import (
	"context"
	"errors"
)

const UnauthorizedUserID = -1

var ErrUserConfict = errors.New("username already registered")

type UserRepository interface {
	Register(ctx context.Context, login string, password string) (userID int, err error)
	Login(ctx context.Context, login string, password string) (userID int, err error)
}

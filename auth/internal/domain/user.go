package domain

import "errors"

const (
	UserProviderEmail  = "email"
	UserProviderGoogle = "google"
	UserProviderYandex = "yandex"
)

type User struct {
	ID       int
	Email    string
	Provider string
}

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrProviderMismatch = errors.New("user provider mismatch")
)

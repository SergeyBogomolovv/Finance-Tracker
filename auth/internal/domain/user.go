package domain

import "errors"

type UserProvider string

const (
	UserProviderEmail  = "email"
	UserProviderGoogle = "google"
	UserProviderYandex = "yandex"
)

type User struct {
	ID              int
	Email           string
	Provider        UserProvider
	IsEmailVerified bool
	AvatarUrl       string
	FullName        string
}

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrProviderMismatch = errors.New("user provider mismatch")
)

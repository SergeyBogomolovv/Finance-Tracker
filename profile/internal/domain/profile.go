package domain

import "errors"

type Profile struct {
	UserID   int
	Email    string
	Provider string
	AvatarID string
	FullName string
}

var (
	ErrProfileNotFound = errors.New("profile not found")
)

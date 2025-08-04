package domain

import (
	"errors"
	"time"
)

type OTP struct {
	UserID    int
	Code      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

var (
	ErrInvalidOTP = errors.New("invalid OTP")
)

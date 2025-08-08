package domain

import (
	"errors"
	"time"
)

type OTP struct {
	ID        int
	Email     string
	Code      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

var (
	ErrInvalidOTP  = errors.New("invalid OTP")
	ErrOTPNotFound = errors.New("OTP not found")
)

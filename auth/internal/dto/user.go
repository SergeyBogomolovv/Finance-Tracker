package dto

import "FinanceTracker/auth/internal/domain"

type CreateUserDto struct {
	Email           string
	Provider        domain.UserProvider
	IsEmailVerified bool
}

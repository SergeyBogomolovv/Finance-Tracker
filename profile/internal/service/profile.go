package service

import (
	"FinanceTracker/profile/internal/domain"
	"FinanceTracker/profile/pkg/transaction"
	"context"
)

type UserRepo interface {
	GetProfileByID(ctx context.Context, userID int) (domain.Profile, error)
}

type profileService struct {
	userRepo  UserRepo
	txManager transaction.Manager
}

func NewProfileService(userRepo UserRepo, txManager transaction.Manager) *profileService {
	return &profileService{
		userRepo:  userRepo,
		txManager: txManager,
	}
}

func (s *profileService) GetProfileInfo(ctx context.Context, userID int) (domain.Profile, error) {
	return s.userRepo.GetProfileByID(ctx, userID)
}

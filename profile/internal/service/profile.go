package service

import (
	"FinanceTracker/profile/internal/domain"
	"FinanceTracker/profile/pkg/events"
	"FinanceTracker/profile/pkg/logger"
	"FinanceTracker/profile/pkg/transaction"
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

type UserRepo interface {
	GetProfileByID(ctx context.Context, userID int) (domain.Profile, error)
	Update(ctx context.Context, user domain.Profile) error
}

type AvatarRepo interface {
	Create(userID int, data io.Reader) (domain.Avatar, error)
}

type profileService struct {
	userRepo   UserRepo
	avatarRepo AvatarRepo
	txManager  transaction.Manager
}

func NewProfileService(userRepo UserRepo, avatarRepo AvatarRepo, txManager transaction.Manager) *profileService {
	return &profileService{
		userRepo:   userRepo,
		txManager:  txManager,
		avatarRepo: avatarRepo,
	}
}

func (s *profileService) GetProfileInfo(ctx context.Context, userID int) (domain.Profile, error) {
	return s.userRepo.GetProfileByID(ctx, userID)
}

func (s *profileService) InitializeUserProfile(ctx context.Context, data events.EventUserRegistered) error {
	return s.txManager.Do(ctx, func(ctx context.Context) error {
		// get profile info
		profile, err := s.userRepo.GetProfileByID(ctx, data.UserID)
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		// set name
		if data.FullName != "" {
			profile.FullName = data.FullName
		} else {
			profile.FullName = generateRandomName()
		}

		var avatar domain.Avatar
		if data.AvatarURL != "" {
			// download avatar
			resp, err := http.Get(data.AvatarURL)
			if err != nil {
				return fmt.Errorf("failed to download avatar: %w", err)
			}
			defer resp.Body.Close()

			// create avatar
			avatar, err = s.avatarRepo.Create(profile.UserID, resp.Body)
			if err != nil {
				return fmt.Errorf("failed to create avatar: %w", err)
			}
			profile.AvatarID = avatar.AvatarID()
		}

		// update db info before upload avatar
		if err := s.userRepo.Update(ctx, profile); err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}

		// upload avatar
		if avatar != nil {
			if err := avatar.Upload(ctx); err != nil {
				return fmt.Errorf("failed to upload avatar: %w", err)
			}
		}

		logger.Debug(ctx, "profile initialized", "user_id", profile.UserID, "full_name", profile.FullName)
		return nil
	})
}

func (s *profileService) UpdateProfile(ctx context.Context, userID int, dto domain.UpdateProfileDto) (domain.Profile, error) {
	// get profile info
	profile, err := s.userRepo.GetProfileByID(ctx, userID)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("failed to get profile: %w", err)
	}

	// update name
	if dto.FullName != nil {
		profile.FullName = *dto.FullName
	}

	var avatar domain.Avatar
	if len(dto.AvatarBytes) > 0 {
		// create avatar
		avatar, err = s.avatarRepo.Create(profile.UserID, bytes.NewReader(dto.AvatarBytes))
		if err != nil {
			return domain.Profile{}, fmt.Errorf("failed to create avatar: %w", err)
		}
		profile.AvatarID = avatar.AvatarID()
	}

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		// update profile
		if err := s.userRepo.Update(ctx, profile); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		// update avatar
		if avatar != nil {
			if err := avatar.Upload(ctx); err != nil {
				return fmt.Errorf("failed to upload avatar: %w", err)
			}
		}

		logger.Debug(ctx, "profile updated", "user_id", userID, "full_name", profile.FullName)
		return nil
	})

	return profile, err
}

var adjectives = []string{
	"Злобный", "Быстрый", "Могучий", "Космический", "Умный", "Веселый",
	"Crazy", "Silent", "Furious", "Lonely", "Brave", "Boyish",
}

var nouns = []string{
	"Волчара", "Медведь", "Космос", "Котяра", "Тигр", "Робот", "Pirate", "Ninja", "Timpany", "Eagle",
}

func generateRandomName() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	return fmt.Sprintf("%s %s", adj, noun)
}

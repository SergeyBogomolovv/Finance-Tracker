package service

import (
	"FinanceTracker/profile/internal/domain"
	"FinanceTracker/profile/pkg/events"
	"FinanceTracker/profile/pkg/logger"
	"FinanceTracker/profile/pkg/transaction"
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
	Upload(ctx context.Context, key string, data io.Reader) error
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
		user, err := s.userRepo.GetProfileByID(ctx, data.UserID)
		if err != nil {
			return err
		}

		if data.FullName != "" {
			user.FullName = data.FullName
		} else {
			user.FullName = generateRandomName()
		}

		if data.AvatarURL != "" {
			user.AvatarID = fmt.Sprintf("%d.jpg", data.UserID)
		} else {
			user.AvatarID = "default.jpg"
		}

		// такой порядок чтобы если будет ошибка загрузки, просто отменилась транзакция и аватарка не была загружена
		if err := s.userRepo.Update(ctx, user); err != nil {
			return err
		}

		if data.AvatarURL != "" {
			resp, err := http.Get(data.AvatarURL)
			if err != nil {
				return fmt.Errorf("failed to download avatar: %w", err)
			}
			defer resp.Body.Close()
			if err := s.avatarRepo.Upload(ctx, user.AvatarID, resp.Body); err != nil {
				return err
			}
		}

		logger.Debug(ctx, "profile initialized", "user_id", data.UserID, "full_name", user.FullName, "avatar_id", user.AvatarID)
		return nil
	})
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

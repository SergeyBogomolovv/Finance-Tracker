package controller

import (
	"FinanceTracker/profile/internal/domain"
	pb "FinanceTracker/profile/pkg/api/profile"
	"FinanceTracker/profile/pkg/logger"
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileService interface {
	GetProfileInfo(ctx context.Context, userID int) (domain.Profile, error)
	UpdateProfile(ctx context.Context, userID int, dto domain.UpdateProfileDto) (domain.Profile, error)
}

type profileController struct {
	pb.UnimplementedProfileServiceServer
	validate *validator.Validate
	svc      ProfileService
}

func NewProfileController(svc ProfileService) *profileController {
	return &profileController{
		svc:      svc,
		validate: validator.New(),
	}
}

func (c *profileController) Register(server *grpc.Server) {
	pb.RegisterProfileServiceServer(server, c)
}

func (c *profileController) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.Profile, error) {
	profile, err := c.svc.GetProfileInfo(ctx, int(req.UserId))
	if errors.Is(err, domain.ErrProfileNotFound) {
		return nil, status.Errorf(codes.NotFound, "profile not found for user ID %d", req.UserId)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get profile for user ID %d: %v", req.UserId, err)
	}

	return &pb.Profile{
		UserId:   int64(profile.UserID),
		Email:    profile.Email,
		Provider: profile.Provider,
		AvatarId: profile.AvatarID,
		FullName: profile.FullName,
	}, nil
}

func (c *profileController) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.Profile, error) {
	profile, err := c.svc.UpdateProfile(ctx, int(req.UserId), domain.UpdateProfileDto{
		FullName:    req.FullName,
		AvatarBytes: req.AvatarBytes,
	})
	if errors.Is(err, domain.ErrProfileNotFound) {
		return nil, status.Errorf(codes.NotFound, "profile not found for user ID %d", req.UserId)
	}
	if err != nil {
		logger.Error(ctx, "failed to update profile", "userID", req.UserId, "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update profile for user ID")
	}

	return &pb.Profile{
		UserId:   int64(profile.UserID),
		Email:    profile.Email,
		Provider: profile.Provider,
		AvatarId: profile.AvatarID,
		FullName: profile.FullName,
	}, nil
}

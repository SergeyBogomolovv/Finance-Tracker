package controller

import (
	"FinanceTracker/profile/internal/domain"
	pb "FinanceTracker/profile/pkg/api/profile"
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileService interface {
	GetProfileInfo(ctx context.Context, userID int) (domain.Profile, error)
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

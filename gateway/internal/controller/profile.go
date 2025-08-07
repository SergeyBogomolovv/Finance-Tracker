package controller

import (
	pb "FinanceTracker/gateway/pkg/api/profile"
	"FinanceTracker/gateway/pkg/logger"
	"FinanceTracker/gateway/pkg/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type profileController struct {
	validate       *validator.Validate
	profileService pb.ProfileServiceClient
	auth           func(http.Handler) http.Handler
}

func NewProfileController(profileService pb.ProfileServiceClient, auth func(http.Handler) http.Handler) *profileController {
	return &profileController{
		validate:       validator.New(),
		profileService: profileService,
		auth:           auth,
	}
}

func (c *profileController) Init(r *http.ServeMux) {
	r.Handle("GET /profile/me", c.auth(http.HandlerFunc(c.handleGetMe)))
}

type ProfileResponse struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
	FullName string `json:"full_name,omitempty"`
	AvatarID string `json:"avatar_id,omitempty"`
}

func protoToProfileResponse(resp *pb.Profile) ProfileResponse {
	return ProfileResponse{
		UserID:   resp.UserId,
		Email:    resp.Email,
		Provider: resp.Provider,
		FullName: resp.FullName,
		AvatarID: resp.AvatarId,
	}
}

// @Summary Получить профиль текущего пользователя
// @Description Возвращает информацию о профиле аутентифицированного пользователя.
// @Tags Профиль
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ProfileResponse "Успешный ответ с данными профиля"
// @Failure 400 {object} utils.ErrorResponse "Профиль не найден"
// @Failure 503 {object} utils.ErrorResponse "Сервис недоступен"
// @Failure 500 {object} utils.ErrorResponse "Внутренняя ошибка сервера"
// @Router /profile/me [get]
func (c *profileController) handleGetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := utils.GetUserID(ctx)

	resp, err := c.profileService.GetProfile(ctx, &pb.GetProfileRequest{
		UserId: userID,
	})

	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				utils.WriteError(w, e.Message(), http.StatusBadRequest)
				return
			case codes.Unavailable:
				logger.Error(ctx, "profile service unavailable", "err", e.Message())
				utils.WriteError(w, "service unavailable", http.StatusServiceUnavailable)
				return
			}
		}

		logger.Error(ctx, "failed to get profile", "err", err)
		utils.WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, protoToProfileResponse(resp), http.StatusOK)
}

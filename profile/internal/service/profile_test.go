package service_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"FinanceTracker/profile/internal/domain"
	"FinanceTracker/profile/internal/service"
	smocks "FinanceTracker/profile/internal/service/mocks"
	"FinanceTracker/profile/pkg/events"
	"FinanceTracker/profile/pkg/logger"
	txmocks "FinanceTracker/profile/pkg/transaction/mocks"
)

func TestProfileService_InitializeUserProfile(t *testing.T) {
	type MockBehavior func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo)

	getErr := errors.New("get error")
	updateErr := errors.New("update error")
	uploadErr := errors.New("upload error")

	base := events.EventUserRegistered{UserID: 77, Email: "u@example.com", Provider: "google"}

	testCases := []struct {
		name            string
		event           events.EventUserRegistered
		mockBehavior    MockBehavior
		wantErr         error
		wantErrContains string
	}{
		{
			name:  "success_with_avatar_and_fullname",
			event: events.EventUserRegistered{UserID: base.UserID, Email: base.Email, Provider: base.Provider, FullName: "John Doe", AvatarURL: ""},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, base.UserID).
					Return(domain.Profile{UserID: base.UserID}, nil)
				users.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
						return p.UserID == base.UserID && p.FullName != "" && (p.AvatarID == "default.jpg" || p.AvatarID == "77.jpg")
					})).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "success_download_and_upload_avatar",
			event: events.EventUserRegistered{UserID: base.UserID, Email: base.Email, Provider: base.Provider, FullName: "", AvatarURL: "http://example/avatar.jpg"},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, base.UserID).
					Return(domain.Profile{UserID: base.UserID}, nil)
				users.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
						return p.UserID == base.UserID && p.FullName != "" && p.AvatarID == "77.jpg"
					})).
					Return(nil)
				avatars.EXPECT().
					Upload(mock.Anything, "77.jpg", mock.Anything).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "get_profile_error",
			event: base,
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, base.UserID).
					Return(domain.Profile{}, getErr)
			},
			wantErr: getErr,
		},
		{
			name:  "update_error",
			event: base,
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, base.UserID).
					Return(domain.Profile{UserID: base.UserID}, nil)
				users.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(updateErr)
			},
			wantErr: updateErr,
		},
		{
			name:  "download_error",
			event: events.EventUserRegistered{UserID: base.UserID, Email: base.Email, Provider: base.Provider, AvatarURL: "http://invalid/404"},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, base.UserID).
					Return(domain.Profile{UserID: base.UserID}, nil)
				users.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErrContains: "failed to download avatar",
		},
		{
			name:  "upload_error",
			event: events.EventUserRegistered{UserID: base.UserID, Email: base.Email, Provider: base.Provider, AvatarURL: "http://example/avatar.jpg"},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, base.UserID).
					Return(domain.Profile{UserID: base.UserID}, nil)
				users.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil)
				avatars.EXPECT().
					Upload(mock.Anything, "77.jpg", mock.Anything).
					Return(uploadErr)
			},
			wantErr: uploadErr,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("avatar-bytes"))
	}))
	defer ts.Close()

	for i := range testCases {
		if testCases[i].name == "success_download_and_upload_avatar" || testCases[i].name == "upload_error" {
			testCases[i].event.AvatarURL = ts.URL + "/avatar.jpg"
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			users := smocks.NewMockUserRepo(t)
			avatars := smocks.NewMockAvatarRepo(t)
			tx := txmocks.NewMockManager(t)

			tx.EXPECT().
				Do(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error { return cb(ctx) })

			if tc.mockBehavior != nil {
				tc.mockBehavior(users, avatars)
			}

			svc := service.NewProfileService(users, avatars, tx)
			ctx := logger.WithLogger(context.Background(), logger.New("test"))
			err := svc.InitializeUserProfile(ctx, tc.event)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			if tc.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrContains)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestProfileService_UpdateProfile(t *testing.T) {
	type MockBehavior func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo)

	getErr := errors.New("get error")
	updateErr := errors.New("update error")
	uploadErr := errors.New("upload error")

	userID := 77
	baseProfile := domain.Profile{UserID: userID, FullName: "Old Name", AvatarID: "old.jpg"}
	newName := "New Name"
	avatarBytes := []byte("avatar-bytes")

	testCases := []struct {
		name            string
		dto             domain.UpdateProfileDto
		mockBehavior    MockBehavior
		wantErr         error
		wantErrContains string
		expectTx        bool
	}{
		{
			name: "success_fullname_only",
			dto:  domain.UpdateProfileDto{FullName: &newName},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, userID).
					Return(baseProfile, nil)
				users.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
						return p.UserID == userID && p.FullName == newName && p.AvatarID == baseProfile.AvatarID
					})).
					Return(nil)
			},
			expectTx: true,
		},
		{
			name: "success_avatar_only",
			dto:  domain.UpdateProfileDto{AvatarBytes: avatarBytes},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, userID).
					Return(baseProfile, nil)
				users.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
						return p.UserID == userID && p.FullName == baseProfile.FullName && p.AvatarID == fmt.Sprintf("%d.jpg", userID)
					})).
					Return(nil)
				avatars.EXPECT().
					Upload(mock.Anything, fmt.Sprintf("%d.jpg", userID), mock.Anything).
					Return(nil)
			},
			expectTx: true,
		},
		{
			name: "success_both",
			dto:  domain.UpdateProfileDto{FullName: &newName, AvatarBytes: avatarBytes},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, userID).
					Return(baseProfile, nil)
				users.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
						return p.UserID == userID && p.FullName == newName && p.AvatarID == fmt.Sprintf("%d.jpg", userID)
					})).
					Return(nil)
				avatars.EXPECT().
					Upload(mock.Anything, fmt.Sprintf("%d.jpg", userID), mock.Anything).
					Return(nil)
			},
			expectTx: true,
		},
		{
			name: "success_noop",
			dto:  domain.UpdateProfileDto{},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, userID).
					Return(baseProfile, nil)
				users.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
						return p.UserID == userID && p.FullName == baseProfile.FullName && p.AvatarID == baseProfile.AvatarID
					})).
					Return(nil)
			},
			expectTx: true,
		},
		{
			name: "get_profile_error",
			dto:  domain.UpdateProfileDto{},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, userID).
					Return(domain.Profile{}, getErr)
			},
			wantErr:  getErr,
			expectTx: false,
		},
		{
			name: "update_error",
			dto:  domain.UpdateProfileDto{FullName: &newName},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, userID).
					Return(baseProfile, nil)
				users.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(updateErr)
			},
			wantErr:  updateErr,
			expectTx: true,
		},
		{
			name: "upload_error",
			dto:  domain.UpdateProfileDto{AvatarBytes: avatarBytes},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().
					GetProfileByID(mock.Anything, userID).
					Return(baseProfile, nil)
				users.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil)
				avatars.EXPECT().
					Upload(mock.Anything, fmt.Sprintf("%d.jpg", userID), mock.Anything).
					Return(uploadErr)
			},
			wantErr:  uploadErr,
			expectTx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			users := smocks.NewMockUserRepo(t)
			avatars := smocks.NewMockAvatarRepo(t)
			tx := txmocks.NewMockManager(t)

			if tc.expectTx {
				tx.EXPECT().
					Do(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error { return cb(ctx) })
			}

			if tc.mockBehavior != nil {
				tc.mockBehavior(users, avatars)
			}

			svc := service.NewProfileService(users, avatars, tx)
			ctx := logger.WithLogger(context.Background(), logger.New("test"))
			got, err := svc.UpdateProfile(ctx, userID, tc.dto)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			if tc.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrContains)
				return
			}
			require.NoError(t, err)

			// Build expected profile based on baseProfile and dto
			expected := baseProfile
			if tc.dto.FullName != nil {
				expected.FullName = *tc.dto.FullName
			}
			if tc.dto.AvatarBytes != nil {
				expected.AvatarID = fmt.Sprintf("%d.jpg", userID)
			}
			assert.Equal(t, expected, got)
		})
	}
}

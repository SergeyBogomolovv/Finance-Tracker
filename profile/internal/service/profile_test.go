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
	dmocks "FinanceTracker/profile/internal/domain/mocks"
	"FinanceTracker/profile/internal/service"
	smocks "FinanceTracker/profile/internal/service/mocks"
	"FinanceTracker/profile/pkg/events"
	"FinanceTracker/profile/pkg/logger"
	txmocks "FinanceTracker/profile/pkg/transaction/mocks"
)

func newAvatarMock(t *testing.T, avatarID string, uploadErr error) domain.Avatar {
	m := dmocks.NewMockAvatar(t)
	m.EXPECT().AvatarID().Return(avatarID)
	m.EXPECT().Upload(mock.Anything).Return(uploadErr)
	return m
}

func TestProfileService_InitializeUserProfile(t *testing.T) {
	type MockBehavior func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo)

	getErr := errors.New("get error")
	updateErr := errors.New("update error")
	uploadErr := errors.New("upload error")

	userID := 77
	avatarID := fmt.Sprintf("avatars/%d.jpg", userID)

	// HTTP server for successful avatar download
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("avatar-bytes"))
	}))
	defer ts.Close()

	base := events.EventUserRegistered{UserID: userID, Email: "u@example.com", Provider: "google"}

	testCases := []struct {
		name            string
		event           events.EventUserRegistered
		mockBehavior    MockBehavior
		wantErr         error
		wantErrContains string
	}{
		{
			name:  "success_with_fullname_without_avatar",
			event: events.EventUserRegistered{UserID: base.UserID, Email: base.Email, Provider: base.Provider, FullName: "John Doe"},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, base.UserID).Return(domain.Profile{UserID: base.UserID}, nil)
				users.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
					return p.UserID == base.UserID && p.FullName == "John Doe" && p.AvatarID == ""
				})).Return(nil)
			},
		},
		{
			name:  "success_download_and_upload_avatar",
			event: events.EventUserRegistered{UserID: base.UserID, Email: base.Email, Provider: base.Provider, AvatarURL: ts.URL + "/avatar.jpg"},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, base.UserID).Return(domain.Profile{UserID: base.UserID}, nil)
				avatars.EXPECT().Create(base.UserID, mock.Anything).Return(newAvatarMock(t, avatarID, nil), nil)
				users.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
					return p.UserID == base.UserID && p.FullName != "" && p.AvatarID == avatarID
				})).Return(nil)
			},
		},
		{
			name:  "get_profile_error",
			event: base,
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, base.UserID).Return(domain.Profile{}, getErr)
			},
			wantErr: getErr,
		},
		{
			name:  "update_error",
			event: base,
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, base.UserID).Return(domain.Profile{UserID: base.UserID}, nil)
				users.EXPECT().Update(mock.Anything, mock.Anything).Return(updateErr)
			},
			wantErr: updateErr,
		},
		{
			name:  "download_error",
			event: events.EventUserRegistered{UserID: base.UserID, Email: base.Email, Provider: base.Provider, AvatarURL: "http://invalid/404"},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, base.UserID).Return(domain.Profile{UserID: base.UserID}, nil)
			},
			wantErrContains: "failed to download avatar",
		},
		{
			name:  "upload_error",
			event: events.EventUserRegistered{UserID: base.UserID, Email: base.Email, Provider: base.Provider, AvatarURL: ts.URL + "/avatar.jpg"},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, base.UserID).Return(domain.Profile{UserID: base.UserID}, nil)
				avatars.EXPECT().Create(base.UserID, mock.Anything).Return(newAvatarMock(t, avatarID, uploadErr), nil)
				users.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
					return p.UserID == base.UserID && p.AvatarID == avatarID
				})).Return(nil)
			},
			wantErr: uploadErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			users := smocks.NewMockUserRepo(t)
			avatars := smocks.NewMockAvatarRepo(t)
			tx := txmocks.NewMockManager(t)

			tx.EXPECT().Do(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error { return cb(ctx) })

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
	avatarID := fmt.Sprintf("avatars/%d.jpg", userID)
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
				users.EXPECT().GetProfileByID(mock.Anything, userID).Return(baseProfile, nil)
				users.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
					return p.UserID == userID && p.FullName == newName && p.AvatarID == baseProfile.AvatarID
				})).Return(nil)
			},
			expectTx: true,
		},
		{
			name: "success_avatar_only",
			dto:  domain.UpdateProfileDto{AvatarBytes: avatarBytes},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, userID).Return(baseProfile, nil)
				avatars.EXPECT().Create(userID, mock.Anything).Return(newAvatarMock(t, avatarID, nil), nil)
				users.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
					return p.UserID == userID && p.FullName == baseProfile.FullName && p.AvatarID == avatarID
				})).Return(nil)
			},
			expectTx: true,
		},
		{
			name: "success_both",
			dto:  domain.UpdateProfileDto{FullName: &newName, AvatarBytes: avatarBytes},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, userID).Return(baseProfile, nil)
				avatars.EXPECT().Create(userID, mock.Anything).Return(newAvatarMock(t, avatarID, nil), nil)
				users.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
					return p.UserID == userID && p.FullName == newName && p.AvatarID == avatarID
				})).Return(nil)
			},
			expectTx: true,
		},
		{
			name: "success_noop",
			dto:  domain.UpdateProfileDto{},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, userID).Return(baseProfile, nil)
				users.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
					return p.UserID == userID && p.FullName == baseProfile.FullName && p.AvatarID == baseProfile.AvatarID
				})).Return(nil)
			},
			expectTx: true,
		},
		{
			name: "get_profile_error",
			dto:  domain.UpdateProfileDto{},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, userID).Return(domain.Profile{}, getErr)
			},
			wantErr:  getErr,
			expectTx: false,
		},
		{
			name: "update_error",
			dto:  domain.UpdateProfileDto{FullName: &newName},
			mockBehavior: func(users *smocks.MockUserRepo, _ *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, userID).Return(baseProfile, nil)
				users.EXPECT().Update(mock.Anything, mock.Anything).Return(updateErr)
			},
			wantErr:  updateErr,
			expectTx: true,
		},
		{
			name: "upload_error",
			dto:  domain.UpdateProfileDto{AvatarBytes: avatarBytes},
			mockBehavior: func(users *smocks.MockUserRepo, avatars *smocks.MockAvatarRepo) {
				users.EXPECT().GetProfileByID(mock.Anything, userID).Return(baseProfile, nil)
				avatars.EXPECT().Create(userID, mock.Anything).Return(newAvatarMock(t, avatarID, uploadErr), nil)
				users.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
					return p.UserID == userID && p.FullName == baseProfile.FullName && p.AvatarID == avatarID
				})).Return(nil)
			},
			wantErr:  uploadErr,
			expectTx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			users := smocks.NewMockUserRepo(t)
			avatars := smocks.NewMockAvatarRepo(t)
			tx := txmocks.NewMockManager(t)

			if tc.expectTx {
				tx.EXPECT().Do(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error { return cb(ctx) })
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

			expected := baseProfile
			if tc.dto.FullName != nil {
				expected.FullName = *tc.dto.FullName
			}
			if tc.dto.AvatarBytes != nil {
				expected.AvatarID = avatarID
			}
			assert.Equal(t, expected, got)
		})
	}
}

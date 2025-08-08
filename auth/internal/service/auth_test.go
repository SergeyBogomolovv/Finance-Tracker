package service_test

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"FinanceTracker/auth/internal/domain"
	"FinanceTracker/auth/internal/dto"
	servicepkg "FinanceTracker/auth/internal/service"
	mocks "FinanceTracker/auth/internal/service/mocks"
	"FinanceTracker/auth/pkg/events"
	"FinanceTracker/auth/pkg/logger"
	txmocks "FinanceTracker/auth/pkg/transaction/mocks"
)

func parseSubjectFromToken(t *testing.T, token string, key []byte) string {
	t.Helper()
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(tk *jwt.Token) (any, error) {
		return key, nil
	})
	require.NoError(t, err)
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	require.True(t, ok)
	return claims.Subject
}

func TestAuthService_OAuth(t *testing.T) {
	type MockBehavior func(users *mocks.MockUserRepo, producer *mocks.MockProducer)

	dbErr := errors.New("db error")
	insertErr := errors.New("insert err")
	kafkaDownErr := errors.New("kafka down")

	testCases := []struct {
		name         string
		payload      dto.OAuthPayload
		mockBehavior MockBehavior
		wantSubj     string
		wantErr      error
	}{
		{
			name: "success_existing_user",
			payload: dto.OAuthPayload{
				Email:    "john@example.com",
				Provider: domain.UserProviderGoogle,
			},
			mockBehavior: func(users *mocks.MockUserRepo, _ *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, "john@example.com").
					Return(domain.User{ID: 10, Email: "john@example.com", Provider: domain.UserProviderGoogle}, nil)
			},
			wantSubj: "10",
			wantErr:  nil,
		},
		{
			name: "success_new_user_registered",
			payload: dto.OAuthPayload{
				Email:     "new@example.com",
				Provider:  domain.UserProviderGoogle,
				FullName:  "New User",
				AvatarUrl: "https://ex.com/a.png",
			},
			mockBehavior: func(users *mocks.MockUserRepo, producer *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, "new@example.com").
					Return(domain.User{}, domain.ErrUserNotFound)

				users.EXPECT().
					Create(mock.Anything, "new@example.com", domain.UserProviderGoogle).
					Return(domain.User{ID: 11, Email: "new@example.com", Provider: domain.UserProviderGoogle}, nil)

				producer.EXPECT().
					PublishUserRegistered(mock.Anything, mock.MatchedBy(func(ev any) bool {
						e, ok := ev.(events.EventUserRegistered)
						if !ok {
							return false
						}
						return e.UserID == 11 && e.Email == "new@example.com" && e.Provider == domain.UserProviderGoogle && e.AvatarURL == "https://ex.com/a.png" && e.FullName == "New User"
					})).
					Return(nil)
			},
			wantSubj: "11",
			wantErr:  nil,
		},
		{
			name: "provider_mismatch",
			payload: dto.OAuthPayload{
				Email:    "mismatch@example.com",
				Provider: domain.UserProviderGoogle,
			},
			mockBehavior: func(users *mocks.MockUserRepo, _ *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, "mismatch@example.com").
					Return(domain.User{ID: 12, Email: "mismatch@example.com", Provider: domain.UserProviderYandex}, nil)
			},
			wantSubj: "",
			wantErr:  domain.ErrProviderMismatch,
		},
		{
			name: "get_user_unknown_error",
			payload: dto.OAuthPayload{
				Email:    "oops@example.com",
				Provider: domain.UserProviderGoogle,
			},
			mockBehavior: func(users *mocks.MockUserRepo, _ *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, "oops@example.com").
					Return(domain.User{}, dbErr)
			},
			wantSubj: "",
			wantErr:  dbErr,
		},
		{
			name: "create_user_error",
			payload: dto.OAuthPayload{
				Email:    "create-fail@example.com",
				Provider: domain.UserProviderGoogle,
			},
			mockBehavior: func(users *mocks.MockUserRepo, _ *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, "create-fail@example.com").
					Return(domain.User{}, domain.ErrUserNotFound)

				users.EXPECT().
					Create(mock.Anything, "create-fail@example.com", domain.UserProviderGoogle).
					Return(domain.User{}, insertErr)
			},
			wantSubj: "",
			wantErr:  insertErr,
		},
		{
			name: "publish_event_error",
			payload: dto.OAuthPayload{
				Email:     "event-fail@example.com",
				Provider:  domain.UserProviderGoogle,
				FullName:  "Event Fail",
				AvatarUrl: "https://ex.com/b.png",
			},
			mockBehavior: func(users *mocks.MockUserRepo, producer *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, "event-fail@example.com").
					Return(domain.User{}, domain.ErrUserNotFound)

				users.EXPECT().
					Create(mock.Anything, "event-fail@example.com", domain.UserProviderGoogle).
					Return(domain.User{ID: 13, Email: "event-fail@example.com", Provider: domain.UserProviderGoogle}, nil)

				producer.EXPECT().
					PublishUserRegistered(mock.Anything, mock.Anything).
					Return(kafkaDownErr)
			},
			wantSubj: "",
			wantErr:  kafkaDownErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := mocks.NewMockUserRepo(t)
			otpRepo := mocks.NewMockOTPRepo(t)
			producer := mocks.NewMockProducer(t)
			txManager := txmocks.NewMockManager(t)

			txManager.EXPECT().
				Do(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
					return cb(ctx)
				})

			if tc.mockBehavior != nil {
				tc.mockBehavior(userRepo, producer)
			}

			jwtKey := []byte("secret")
			svc := servicepkg.NewAuthService(userRepo, otpRepo, producer, txManager, time.Minute, jwtKey)

			ctx := logger.WithLogger(context.Background(), logger.New("test"))
			gotToken, err := svc.OAuth(ctx, tc.payload)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, gotToken)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, gotToken)
			subj := parseSubjectFromToken(t, gotToken, jwtKey)
			assert.Equal(t, tc.wantSubj, subj)

			if tc.wantSubj != "" {
				_, convErr := strconv.Atoi(subj)
				assert.NoError(t, convErr)
			}
		})
	}
}

func TestAuthService_GenerateOTP(t *testing.T) {
	type MockBehavior func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, producer *mocks.MockProducer)

	dbErr := errors.New("db error")
	genErr := errors.New("gen error")
	kafkaErr := errors.New("kafka error")

	email := "user@example.com"
	duration := 5 * time.Minute
	now := time.Now().UTC()
	otp := domain.OTP{Email: email, Code: "123456", CreatedAt: now, ExpiresAt: now.Add(duration)}

	testCases := []struct {
		name         string
		mockBehavior MockBehavior
		wantErr      error
	}{
		{
			name: "success_user_not_found",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, producer *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{}, domain.ErrUserNotFound)

				otps.EXPECT().
					Generate(mock.Anything, email, mock.MatchedBy(func(d time.Duration) bool { return d == duration })).
					Return(otp, nil)

				producer.EXPECT().
					PublishOTPGenerated(mock.Anything, mock.MatchedBy(func(ev any) bool {
						e, ok := ev.(events.EventOTPGenerated)
						if !ok {
							return false
						}
						return e.Email == otp.Email && e.Code == otp.Code && e.CreatedAt.Equal(otp.CreatedAt) && e.ExpiresAt.Equal(otp.ExpiresAt)
					})).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "success_existing_email_provider",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, producer *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{ID: 1, Email: email, Provider: domain.UserProviderEmail}, nil)

				otps.EXPECT().
					Generate(mock.Anything, email, mock.MatchedBy(func(d time.Duration) bool { return d == duration })).
					Return(otp, nil)

				producer.EXPECT().
					PublishOTPGenerated(mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "provider_mismatch",
			mockBehavior: func(users *mocks.MockUserRepo, _ *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{ID: 2, Email: email, Provider: domain.UserProviderGoogle}, nil)
			},
			wantErr: domain.ErrProviderMismatch,
		},
		{
			name: "get_user_unknown_error",
			mockBehavior: func(users *mocks.MockUserRepo, _ *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{}, dbErr)
			},
			wantErr: dbErr,
		},
		{
			name: "generate_error",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{}, domain.ErrUserNotFound)

				otps.EXPECT().
					Generate(mock.Anything, email, mock.Anything).
					Return(domain.OTP{}, genErr)
			},
			wantErr: genErr,
		},
		{
			name: "publish_error",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, producer *mocks.MockProducer) {
				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{}, domain.ErrUserNotFound)

				otps.EXPECT().
					Generate(mock.Anything, email, mock.Anything).
					Return(otp, nil)

				producer.EXPECT().
					PublishOTPGenerated(mock.Anything, mock.Anything).
					Return(kafkaErr)
			},
			wantErr: kafkaErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := mocks.NewMockUserRepo(t)
			otpRepo := mocks.NewMockOTPRepo(t)
			producer := mocks.NewMockProducer(t)
			txManager := txmocks.NewMockManager(t)

			txManager.EXPECT().
				Do(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error { return cb(ctx) })

			if tc.mockBehavior != nil {
				tc.mockBehavior(userRepo, otpRepo, producer)
			}

			svc := servicepkg.NewAuthService(userRepo, otpRepo, producer, txManager, time.Minute, []byte("secret"))
			ctx := logger.WithLogger(context.Background(), logger.New("test"))
			err := svc.GenerateOTP(ctx, email)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAuthService_VerifyOTP(t *testing.T) {
	type MockBehavior func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, producer *mocks.MockProducer)

	email := "user@example.com"
	code := "123456"

	verifyErr := errors.New("verify err")
	markErr := errors.New("mark err")
	getErr := errors.New("get err")
	createErr := errors.New("create err")
	publishErr := errors.New("publish err")

	testCases := []struct {
		name         string
		mockBehavior MockBehavior
		wantSubj     string
		wantErr      error
	}{
		{
			name: "success_existing_user",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(true, nil)

				otps.EXPECT().
					MarkUsed(mock.Anything, email, code).
					Return(nil)

				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{ID: 21, Email: email, Provider: domain.UserProviderEmail}, nil)
			},
			wantSubj: "21",
			wantErr:  nil,
		},
		{
			name: "success_new_user_registered",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, producer *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(true, nil)

				otps.EXPECT().
					MarkUsed(mock.Anything, email, code).
					Return(nil)

				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{}, domain.ErrUserNotFound)

				users.EXPECT().
					Create(mock.Anything, email, domain.UserProviderEmail).
					Return(domain.User{ID: 22, Email: email, Provider: domain.UserProviderEmail}, nil)

				producer.EXPECT().
					PublishUserRegistered(mock.Anything, mock.MatchedBy(func(ev any) bool {
						e, ok := ev.(events.EventUserRegistered)
						if !ok {
							return false
						}
						return e.UserID == 22 && e.Email == email && e.Provider == domain.UserProviderEmail
					})).
					Return(nil)
			},
			wantSubj: "22",
			wantErr:  nil,
		},
		{
			name: "verify_error",
			mockBehavior: func(_ *mocks.MockUserRepo, otps *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(false, verifyErr)
			},
			wantSubj: "",
			wantErr:  verifyErr,
		},
		{
			name: "invalid_otp",
			mockBehavior: func(_ *mocks.MockUserRepo, otps *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(false, nil)
			},
			wantSubj: "",
			wantErr:  domain.ErrInvalidOTP,
		},
		{
			name: "mark_used_error",
			mockBehavior: func(_ *mocks.MockUserRepo, otps *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(true, nil)

				otps.EXPECT().
					MarkUsed(mock.Anything, email, code).
					Return(markErr)
			},
			wantSubj: "",
			wantErr:  markErr,
		},
		{
			name: "get_user_error",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(true, nil)

				otps.EXPECT().
					MarkUsed(mock.Anything, email, code).
					Return(nil)

				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{}, getErr)
			},
			wantSubj: "",
			wantErr:  getErr,
		},
		{
			name: "provider_mismatch",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(true, nil)

				otps.EXPECT().
					MarkUsed(mock.Anything, email, code).
					Return(nil)

				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{ID: 25, Email: email, Provider: domain.UserProviderGoogle}, nil)
			},
			wantSubj: "",
			wantErr:  domain.ErrProviderMismatch,
		},
		{
			name: "create_user_error",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, _ *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(true, nil)

				otps.EXPECT().
					MarkUsed(mock.Anything, email, code).
					Return(nil)

				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{}, domain.ErrUserNotFound)

				users.EXPECT().
					Create(mock.Anything, email, domain.UserProviderEmail).
					Return(domain.User{}, createErr)
			},
			wantSubj: "",
			wantErr:  createErr,
		},
		{
			name: "publish_event_error",
			mockBehavior: func(users *mocks.MockUserRepo, otps *mocks.MockOTPRepo, producer *mocks.MockProducer) {
				otps.EXPECT().
					Verify(mock.Anything, email, code).
					Return(true, nil)

				otps.EXPECT().
					MarkUsed(mock.Anything, email, code).
					Return(nil)

				users.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(domain.User{}, domain.ErrUserNotFound)

				users.EXPECT().
					Create(mock.Anything, email, domain.UserProviderEmail).
					Return(domain.User{ID: 26, Email: email, Provider: domain.UserProviderEmail}, nil)

				producer.EXPECT().
					PublishUserRegistered(mock.Anything, mock.Anything).
					Return(publishErr)
			},
			wantSubj: "",
			wantErr:  publishErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := mocks.NewMockUserRepo(t)
			otpRepo := mocks.NewMockOTPRepo(t)
			producer := mocks.NewMockProducer(t)
			txManager := txmocks.NewMockManager(t)

			txManager.EXPECT().
				Do(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error { return cb(ctx) })

			if tc.mockBehavior != nil {
				tc.mockBehavior(userRepo, otpRepo, producer)
			}

			jwtKey := []byte("secret")
			svc := servicepkg.NewAuthService(userRepo, otpRepo, producer, txManager, time.Minute, jwtKey)
			ctx := logger.WithLogger(context.Background(), logger.New("test"))
			gotToken, err := svc.VerifyOTP(ctx, email, code)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, gotToken)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, gotToken)
			subj := parseSubjectFromToken(t, gotToken, jwtKey)
			assert.Equal(t, tc.wantSubj, subj)
		})
	}
}

package repo

import (
	"FinanceTracker/auth/internal/domain"
	"FinanceTracker/auth/pkg/transaction"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type OTP struct {
	UserID    int       `db:"user_id"`
	Code      string    `db:"code"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}

type otpRepo struct {
	storage *sqlx.DB
	qb      sq.StatementBuilderType
}

func NewOTPRepo(storage *sqlx.DB) *otpRepo {
	return &otpRepo{
		storage: storage,
		qb:      sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *otpRepo) Generate(ctx context.Context, userID int) (domain.OTP, error) {
	code, err := generateCode()
	if err != nil {
		return domain.OTP{}, fmt.Errorf("failed to generate OTP code: %w", err)
	}

	otp := OTP{
		UserID:    userID,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute), // OTP valid for 5 minutes
	}

	query, args := r.qb.Insert("otps").
		Columns("user_id", "code", "created_at", "expires_at").
		Values(otp.UserID, otp.Code, otp.CreatedAt, otp.ExpiresAt).
		Suffix("RETURNING *").MustSql()

	if err := r.getContext(ctx, &otp, query, args...); err != nil {
		return domain.OTP{}, fmt.Errorf("failed to insert OTP: %w", err)
	}

	return domain.OTP{
		UserID:    otp.UserID,
		Code:      otp.Code,
		CreatedAt: otp.CreatedAt,
		ExpiresAt: otp.ExpiresAt,
	}, nil
}

func (r *otpRepo) Validate(ctx context.Context, userID int, code string) (bool, error) {
	query, args := r.qb.Select("TRUE").
		From("otps").
		Where(sq.Eq{"user_id": userID, "code": code}).
		Where("expires_at > NOW()").
		MustSql()
	var valid bool
	if err := r.getContext(ctx, &valid, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to validate OTP: %w", err)
	}
	return true, nil
}

func (r *otpRepo) DeleteAll(ctx context.Context, userID int) error {
	query, args := r.qb.Delete("otps").
		Where(sq.Eq{"user_id": userID}).
		MustSql()

	return r.execContext(ctx, query, args...)
}

func generateCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (r *otpRepo) execContext(ctx context.Context, query string, args ...any) error {
	tx := transaction.ExtractTx(ctx)
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, args...)
		return err
	}
	_, err := r.storage.ExecContext(ctx, query, args...)
	return err
}

func (r *otpRepo) getContext(ctx context.Context, dest any, query string, args ...any) error {
	tx := transaction.ExtractTx(ctx)
	if tx != nil {
		return tx.GetContext(ctx, dest, query, args...)
	}
	return r.storage.GetContext(ctx, dest, query, args...)
}

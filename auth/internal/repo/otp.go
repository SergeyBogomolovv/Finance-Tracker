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
	ID        int       `db:"otp_id"`
	Email     string    `db:"email"`
	Code      string    `db:"code"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
	IsUsed    bool      `db:"is_used"`
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

func (r *otpRepo) Generate(ctx context.Context, email string, duration time.Duration) (domain.OTP, error) {
	code, err := generateCode()
	if err != nil {
		return domain.OTP{}, fmt.Errorf("failed to generate OTP code: %w", err)
	}

	var otp OTP
	query, args := r.qb.
		Insert("email_otps").
		Columns("email", "code", "created_at", "expires_at").
		Values(email, code, time.Now(), time.Now().Add(duration)).
		Suffix("RETURNING otp_id, email, code, created_at, expires_at").
		MustSql()

	if err := r.getContext(ctx, &otp, query, args...); err != nil {
		return domain.OTP{}, fmt.Errorf("failed to insert OTP: %w", err)
	}

	return domain.OTP{
		ID:        otp.ID,
		Email:     otp.Email,
		Code:      otp.Code,
		CreatedAt: otp.CreatedAt,
		ExpiresAt: otp.ExpiresAt,
	}, nil
}

func (r *otpRepo) Verify(ctx context.Context, email, code string) (bool, error) {
	query, args := r.qb.
		Select("TRUE").
		From("email_otps").
		Where(sq.Eq{"email": email, "code": code, "is_used": false}).
		Where(sq.Gt{"expires_at": time.Now()}).
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

func (r *otpRepo) MarkUsed(ctx context.Context, email, code string) error {
	query, args := r.qb.
		Update("email_otps").
		Set("is_used", true).
		Where(sq.Eq{"email": email, "code": code}).
		MustSql()

	aff, err := r.execContext(ctx, query, args...)
	if err != nil {
		return err
	}
	if aff == 0 {
		return domain.ErrOTPNotFound
	}
	return nil
}

// func (r *otpRepo) DeleteAll(ctx context.Context, userID int) error {
// 	query, args := r.qb.Delete("otps").
// 		Where(sq.Eq{"user_id": userID}).
// 		MustSql()

// 	_, err := r.execContext(ctx, query, args...)
// 	return err
// }

func generateCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (r *otpRepo) execContext(ctx context.Context, query string, args ...any) (int64, error) {
	tx := transaction.ExtractTx(ctx)
	if tx != nil {
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}

	res, err := r.storage.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (r *otpRepo) getContext(ctx context.Context, dest any, query string, args ...any) error {
	tx := transaction.ExtractTx(ctx)
	if tx != nil {
		return tx.GetContext(ctx, dest, query, args...)
	}
	return r.storage.GetContext(ctx, dest, query, args...)
}

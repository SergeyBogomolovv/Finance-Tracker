package repo

import (
	"FinanceTracker/profile/internal/domain"
	"FinanceTracker/profile/pkg/transaction"
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID        int            `db:"user_id"`
	Email     string         `db:"email"`
	Provider  string         `db:"provider"`
	FullName  sql.NullString `db:"full_name"`
	AvatarID  sql.NullString `db:"avatar_id"`
	CreatedAt time.Time      `db:"created_at"`
}

func (u User) ToProfile() domain.Profile {
	return domain.Profile{
		UserID:   u.ID,
		Email:    u.Email,
		Provider: u.Provider,
		AvatarID: u.AvatarID.String,
		FullName: u.FullName.String,
	}
}

type userRepo struct {
	storage *sqlx.DB
	qb      sq.StatementBuilderType
}

func NewUserRepo(storage *sqlx.DB) *userRepo {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return &userRepo{
		storage: storage,
		qb:      qb,
	}
}

func (r *userRepo) GetProfileByID(ctx context.Context, userID int) (domain.Profile, error) {
	query, args := r.qb.Select("user_id", "email", "provider", "full_name", "avatar_id", "created_at").
		From("users").
		Where(sq.Eq{"user_id": userID}).
		MustSql()

	var user User
	err := r.getContext(ctx, &user, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	if err != nil {
		return domain.Profile{}, err
	}

	return user.ToProfile(), nil
}

func (r *userRepo) Update(ctx context.Context, user domain.Profile) error {
	m := make(map[string]any)
	if user.FullName != "" {
		m["full_name"] = user.FullName
	}
	if user.AvatarID != "" {
		m["avatar_id"] = user.AvatarID
	}
	query, args := r.qb.Update("users").SetMap(m).Where(sq.Eq{"user_id": user.UserID}).MustSql()
	res, err := r.execContext(ctx, query, args...)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return domain.ErrProfileNotFound
	}
	return nil
}

func (r *userRepo) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx := transaction.ExtractTx(ctx)
	if tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return r.storage.ExecContext(ctx, query, args...)
}

func (r *userRepo) getContext(ctx context.Context, dest any, query string, args ...any) error {
	tx := transaction.ExtractTx(ctx)
	if tx != nil {
		return tx.GetContext(ctx, dest, query, args...)
	}
	return r.storage.GetContext(ctx, dest, query, args...)
}

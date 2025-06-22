package repo

import (
	"FinanceTracker/auth/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID              int            `db:"user_id"`
	Email           string         `db:"email"`
	Provider        string         `db:"provider"`
	IsEmailVerified bool           `db:"is_email_verified"`
	FullName        sql.NullString `db:"full_name"`
	AvatarUrl       sql.NullString `db:"avatar_url"`
	CreatedAt       time.Time      `db:"created_at"`
}

func (u User) ToDomain() domain.User {
	return domain.User{
		ID:              u.ID,
		Email:           u.Email,
		Provider:        domain.UserProvider(u.Provider),
		IsEmailVerified: u.IsEmailVerified,
		AvatarUrl:       u.AvatarUrl.String,
		FullName:        u.FullName.String,
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

func (r *userRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query, args := r.qb.Select("user_id", "email", "provider", "is_email_verified", "full_name", "avatar_url", "created_at").
		From("users").
		Where(sq.Eq{"email": email}).
		MustSql()

	var user User
	err := r.storage.GetContext(ctx, &user, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return user.ToDomain(), nil
}

func (r *userRepo) Create(ctx context.Context, user domain.User) (domain.User, error) {
	m := map[string]any{
		"email":             user.Email,
		"provider":          string(user.Provider),
		"is_email_verified": user.IsEmailVerified,
	}
	if user.FullName != "" {
		m["full_name"] = user.FullName
	}
	if user.AvatarUrl != "" {
		m["avatar_url"] = user.AvatarUrl
	}

	query, args := r.qb.Insert("users").SetMap(m).
		Suffix("RETURNING user_id, email, provider, is_email_verified, full_name, avatar_url, created_at").
		MustSql()

	var createdUser User
	err := r.storage.GetContext(ctx, &createdUser, query, args...)
	if err != nil {
		return domain.User{}, err
	}

	return createdUser.ToDomain(), nil
}

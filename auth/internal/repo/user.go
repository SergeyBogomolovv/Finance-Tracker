package repo

import (
	"FinanceTracker/auth/internal/domain"
	"FinanceTracker/auth/pkg/transaction"
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID        int       `db:"user_id"`
	Email     string    `db:"email"`
	Provider  string    `db:"provider"`
	CreatedAt time.Time `db:"created_at"`
}

func (u User) ToDomain() domain.User {
	return domain.User{
		ID:       u.ID,
		Email:    u.Email,
		Provider: u.Provider,
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
	query, args := r.qb.Select("user_id", "email", "provider", "created_at").
		From("users").
		Where(sq.Eq{"email": email}).
		MustSql()

	var user User
	err := r.getContext(ctx, &user, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return user.ToDomain(), nil
}

func (r *userRepo) GetByID(ctx context.Context, userID int) (domain.User, error) {
	query, args := r.qb.Select("user_id", "email", "provider", "created_at").
		From("users").
		Where(sq.Eq{"user_id": userID}).
		MustSql()

	var user User
	err := r.getContext(ctx, &user, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return user.ToDomain(), nil
}

func (r *userRepo) Create(ctx context.Context, email, provider string) (domain.User, error) {
	query, args := r.qb.Insert("users").
		Columns("email", "provider").
		Values(email, provider).
		Suffix("RETURNING user_id, email, provider, created_at").
		MustSql()

	var createdUser User
	err := r.getContext(ctx, &createdUser, query, args...)
	if err != nil {
		return domain.User{}, err
	}

	return createdUser.ToDomain(), nil
}

func (r *userRepo) getContext(ctx context.Context, dest any, query string, args ...any) error {
	tx := transaction.ExtractTx(ctx)
	if tx != nil {
		return tx.GetContext(ctx, dest, query, args...)
	}
	return r.storage.GetContext(ctx, dest, query, args...)
}

package repo

import (
	"FinanceTracker/notification/internal/domain"
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

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

func (r *userRepo) GetEmailByID(ctx context.Context, userID int) (string, error) {
	query, args := r.qb.Select("email").
		From("users").
		Where(sq.Eq{"user_id": userID}).
		MustSql()

	var email string
	err := r.storage.GetContext(ctx, &email, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrUserNotFound
	}
	if err != nil {
		return "", err
	}

	return email, nil
}

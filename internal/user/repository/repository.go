package repository

import (
	"context"
	"database/sql"
	"errors"

	"family-shopping-list-api/internal/user/model"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user *model.User) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, nickname) VALUES (?, ?, ?)`,
		user.Username, user.PasswordHash, user.Nickname,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = uint64(id)
	return nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, nickname, password_hash, created_at, updated_at
		 FROM users WHERE username = ?`, username,
	).Scan(&user.ID, &user.Username, &user.Nickname, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, nickname, password_hash, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Username, &user.Nickname, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

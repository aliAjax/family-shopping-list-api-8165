package repository

import (
	"context"
	"database/sql"

	"family-shopping-list-api/internal/member/model"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Add(ctx context.Context, listID, userID uint64, role string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO members (list_id, user_id, role, joined_at)
		 VALUES (?, ?, ?, UTC_TIMESTAMP())
		 ON DUPLICATE KEY UPDATE role = VALUES(role), updated_at = UTC_TIMESTAMP()`,
		listID, userID, role,
	)
	return err
}

func (r *Repository) IsMember(ctx context.Context, listID, userID uint64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM members WHERE list_id = ? AND user_id = ?`,
		listID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *Repository) IsOwner(ctx context.Context, listID, userID uint64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM members WHERE list_id = ? AND user_id = ? AND role = 'owner'`,
		listID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *Repository) List(ctx context.Context, listID uint64) ([]model.Member, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT m.id, m.list_id, m.user_id, m.role, u.username, u.nickname,
		        m.joined_at, m.created_at, m.updated_at
		 FROM members m
		 JOIN users u ON u.id = m.user_id
		 WHERE m.list_id = ?
		 ORDER BY m.role = 'owner' DESC, m.joined_at ASC`, listID,
	)
	if err != nil {
		return nil, err
	}
	members := make([]model.Member, 0)
	for rows.Next() {
		var m model.Member
		if err := rows.Scan(&m.ID, &m.ListID, &m.UserID, &m.Role, &m.Username, &m.Nickname,
			&m.JoinedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repository) GetByUser(ctx context.Context, listID, userID uint64) (*model.Member, error) {
	var m model.Member
	err := r.db.QueryRowContext(ctx,
		`SELECT m.id, m.list_id, m.user_id, m.role, u.username, u.nickname,
		        m.joined_at, m.created_at, m.updated_at
		 FROM members m
		 JOIN users u ON u.id = m.user_id
		 WHERE m.list_id = ? AND m.user_id = ?`,
		listID, userID,
	).Scan(&m.ID, &m.ListID, &m.UserID, &m.Role, &m.Username, &m.Nickname,
		&m.JoinedAt, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetByID(ctx context.Context, listID, memberID uint64) (*model.Member, error) {
	var m model.Member
	err := r.db.QueryRowContext(ctx,
		`SELECT m.id, m.list_id, m.user_id, m.role, u.username, u.nickname,
		        m.joined_at, m.created_at, m.updated_at
		 FROM members m
		 JOIN users u ON u.id = m.user_id
		 WHERE m.id = ? AND m.list_id = ?`,
		memberID, listID,
	).Scan(&m.ID, &m.ListID, &m.UserID, &m.Role, &m.Username, &m.Nickname,
		&m.JoinedAt, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) Remove(ctx context.Context, listID, memberID uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM members WHERE id = ? AND list_id = ? AND role <> 'owner'`, memberID, listID)
	return err
}

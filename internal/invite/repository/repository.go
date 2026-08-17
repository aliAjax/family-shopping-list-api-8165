package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"family-shopping-list-api/internal/invite/model"
)

var ErrNotFound = errors.New("邀请码不存在或已失效")

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, invite *model.Invite) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO invites (list_id, code, created_by, max_uses, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		invite.ListID, invite.Code, invite.CreatedBy, invite.MaxUses, invite.ExpiresAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	invite.ID = uint64(id)
	return nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*model.Invite, error) {
	var invite model.Invite
	err := r.db.QueryRowContext(ctx,
		`SELECT i.id, i.list_id, i.code, i.created_by, u.nickname,
		        i.max_uses, i.used_count, i.expires_at, i.active, i.created_at, i.updated_at
		 FROM invites i
		 JOIN users u ON u.id = i.created_by
		 WHERE i.code = ?`, code,
	).Scan(&invite.ID, &invite.ListID, &invite.Code, &invite.CreatedBy, &invite.CreatorName,
		&invite.MaxUses, &invite.UsedCount, &invite.ExpiresAt, &invite.Active, &invite.CreatedAt, &invite.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询邀请码失败: %w", err)
	}
	return &invite, nil
}

func (r *Repository) ListByList(ctx context.Context, listID uint64) ([]model.Invite, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT i.id, i.list_id, i.code, i.created_by, u.nickname,
		        i.max_uses, i.used_count, i.expires_at, i.active, i.created_at, i.updated_at
		 FROM invites i
		 JOIN users u ON u.id = i.created_by
		 WHERE i.list_id = ?
		 ORDER BY i.created_at DESC`, listID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invites := make([]model.Invite, 0)
	for rows.Next() {
		var invite model.Invite
		if err := rows.Scan(&invite.ID, &invite.ListID, &invite.Code, &invite.CreatedBy, &invite.CreatorName,
			&invite.MaxUses, &invite.UsedCount, &invite.ExpiresAt, &invite.Active, &invite.CreatedAt, &invite.UpdatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, rows.Err()
}

func (r *Repository) IncrementUsed(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE invites SET used_count = used_count + 1, updated_at = UTC_TIMESTAMP() WHERE id = ?`, id)
	return err
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"family-shopping-list-api/internal/list/model"
)

var ErrNotFound = errors.New("清单不存在")

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, list *model.ShoppingList) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO shopping_lists (name, description, owner_id) VALUES (?, ?, ?)`,
		list.Name, list.Description, list.OwnerID,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	list.ID = uint64(id)
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*model.ShoppingList, error) {
	var list model.ShoppingList
	err := r.db.QueryRowContext(ctx,
		`SELECT l.id, l.name, l.description, l.owner_id, u.nickname,
		        (SELECT COUNT(*) FROM members m WHERE m.list_id = l.id) AS member_count,
		        l.created_at, l.updated_at
		 FROM shopping_lists l
		 JOIN users u ON u.id = l.owner_id
		 WHERE l.id = ?`, id,
	).Scan(&list.ID, &list.Name, &list.Description, &list.OwnerID, &list.OwnerName,
		&list.MemberCount, &list.CreatedAt, &list.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询清单失败: %w", err)
	}
	return &list, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uint64) ([]model.ShoppingList, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT l.id, l.name, l.description, l.owner_id, u.nickname,
		        (SELECT COUNT(*) FROM members m WHERE m.list_id = l.id) AS member_count,
		        l.created_at, l.updated_at
		 FROM shopping_lists l
		 JOIN users u ON u.id = l.owner_id
		 JOIN members me ON me.list_id = l.id AND me.user_id = ?
		 ORDER BY l.updated_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lists := make([]model.ShoppingList, 0)
	for rows.Next() {
		var list model.ShoppingList
		if err := rows.Scan(&list.ID, &list.Name, &list.Description, &list.OwnerID, &list.OwnerName,
			&list.MemberCount, &list.CreatedAt, &list.UpdatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, rows.Err()
}

func (r *Repository) Update(ctx context.Context, list *model.ShoppingList) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE shopping_lists SET name = ?, description = ?, updated_at = UTC_TIMESTAMP() WHERE id = ?`,
		list.Name, list.Description, list.ID,
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM shopping_lists WHERE id = ?`, id)
	return err
}

func (r *Repository) GetByIDForUpdate(ctx context.Context, id uint64) (*model.ShoppingList, error) {
	// Kept separate so callers can read basic list data without aggregate fields if needed.
	return r.GetByID(ctx, id)
}

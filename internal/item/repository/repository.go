package repository

import (
	"context"
	"database/sql"
	"errors"

	"family-shopping-list-api/internal/item/model"
)

var ErrNotFound = errors.New("商品不存在")

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, item *model.Item) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO items (list_id, name, quantity, purchased, created_by, updated_by)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		item.ListID, item.Name, item.Quantity, item.Purchased, item.CreatedBy, item.UpdatedBy,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = uint64(id)
	return nil
}

func (r *Repository) GetByID(ctx context.Context, listID, itemID uint64) (*model.Item, error) {
	var item model.Item
	err := r.db.QueryRowContext(ctx,
		`SELECT i.id, i.list_id, i.name, i.quantity, i.purchased,
		        i.created_by, i.updated_by, cu.nickname, uu.nickname,
		        i.created_at, i.updated_at
		 FROM items i
		 JOIN users cu ON cu.id = i.created_by
		 JOIN users uu ON uu.id = i.updated_by
		 WHERE i.id = ? AND i.list_id = ?`, itemID, listID,
	).Scan(&item.ID, &item.ListID, &item.Name, &item.Quantity, &item.Purchased,
		&item.CreatedBy, &item.UpdatedBy, &item.CreatedByName, &item.UpdatedByName,
		&item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListByList(ctx context.Context, listID uint64) ([]model.Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT i.id, i.list_id, i.name, i.quantity, i.purchased,
		        i.created_by, i.updated_by, cu.nickname, uu.nickname,
		        i.created_at, i.updated_at
		 FROM items i
		 JOIN users cu ON cu.id = i.created_by
		 JOIN users uu ON uu.id = i.updated_by
		 WHERE i.list_id = ?
		 ORDER BY i.created_at DESC, i.id DESC`, listID,
	)
	if err != nil {
		return nil, err
	}
	items := make([]model.Item, 0)
	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.ID, &item.ListID, &item.Name, &item.Quantity, &item.Purchased,
			&item.CreatedBy, &item.UpdatedBy, &item.CreatedByName, &item.UpdatedByName,
			&item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Update(ctx context.Context, item *model.Item) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE items SET name = ?, quantity = ?, purchased = ?, updated_by = ?, updated_at = UTC_TIMESTAMP()
		 WHERE id = ? AND list_id = ?`,
		item.Name, item.Quantity, item.Purchased, item.UpdatedBy, item.ID, item.ListID,
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, listID, itemID uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM items WHERE id = ? AND list_id = ?`, itemID, listID)
	return err
}

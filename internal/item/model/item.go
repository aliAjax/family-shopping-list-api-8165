package model

import "time"

type Item struct {
	ID            uint64    `json:"id"`
	ListID        uint64    `json:"list_id"`
	Name          string    `json:"name"`
	Quantity      int       `json:"quantity"`
	Purchased     bool      `json:"purchased"`
	CreatedBy     uint64    `json:"created_by"`
	UpdatedBy     uint64    `json:"updated_by"`
	CreatedByName string    `json:"created_by_name"`
	UpdatedByName string    `json:"last_modified_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"last_modified_at"`
}

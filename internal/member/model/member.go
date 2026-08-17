package model

import "time"

type Member struct {
	ID        uint64    `json:"id"`
	ListID    uint64    `json:"list_id"`
	UserID    uint64    `json:"user_id"`
	Role      string    `json:"role"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	JoinedAt  time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package model

import "time"

type User struct {
	ID           uint64    `json:"id"`
	Username     string    `json:"username"`
	Nickname     string    `json:"nickname"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

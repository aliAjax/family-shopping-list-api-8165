package model

import "time"

type Invite struct {
	ID          uint64     `json:"id"`
	ListID      uint64     `json:"list_id"`
	Code        string     `json:"code"`
	CreatedBy   uint64     `json:"created_by"`
	CreatorName string     `json:"creator_name,omitempty"`
	MaxUses     int        `json:"max_uses"`
	UsedCount   int        `json:"used_count"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

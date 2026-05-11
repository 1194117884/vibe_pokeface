package model

import (
	"context"
	"time"
)

type User struct {
	ID        int64     `db:"id" json:"id"`
	Nickname  string    `db:"nickname" json:"nickname"`
	AvatarURL *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	Role      string    `db:"role" json:"role"`
	Status    int8      `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type UserAuth struct {
	ID          int64  `db:"id"`
	UserID      int64  `db:"user_id"`
	Provider    string `db:"provider"`
	ProviderUID string `db:"provider_uid"`
	Credential  string `db:"credential"`
}

type UserStore interface {
	Create(ctx context.Context, user *User) error
	FindByProvider(ctx context.Context, provider, providerUID string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	CreateAuth(ctx context.Context, auth *UserAuth) error
	FindAuth(ctx context.Context, provider, providerUID string) (*UserAuth, error)
	// Admin methods
	ListUsers(ctx context.Context, offset, limit int) ([]User, error)
	SearchUsers(ctx context.Context, query string) ([]User, error)
	UpdateUserStatus(ctx context.Context, userID int64, status int8) error
	GetUserCount(ctx context.Context) (int, error)
}

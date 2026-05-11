package model

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type UserDB struct {
	db *sqlx.DB
}

func NewUserDB(db *sqlx.DB) *UserDB {
	return &UserDB{db: db}
}

func (s *UserDB) Create(ctx context.Context, user *User) error {
	result, err := s.db.ExecContext(ctx, "INSERT INTO users (nickname, role) VALUES (?, ?)", user.Nickname, user.Role)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	user.ID = id
	return nil
}

func (s *UserDB) FindByProvider(ctx context.Context, provider, providerUID string) (*User, error) {
	var user User
	err := s.db.GetContext(ctx, &user, `
        SELECT u.* FROM users u
        JOIN user_auths ua ON u.id = ua.user_id
        WHERE ua.provider = ? AND ua.provider_uid = ?
    `, provider, providerUID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserDB) FindByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := s.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserDB) CreateAuth(ctx context.Context, ua *UserAuth) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO user_auths (user_id, provider, provider_uid, credential) VALUES (?, ?, ?, ?)",
		ua.UserID, ua.Provider, ua.ProviderUID, ua.Credential)
	return err
}

func (s *UserDB) FindAuth(ctx context.Context, provider, providerUID string) (*UserAuth, error) {
	var ua UserAuth
	err := s.db.GetContext(ctx, &ua, "SELECT * FROM user_auths WHERE provider = ? AND provider_uid = ?", provider, providerUID)
	if err != nil {
		return nil, err
	}
	return &ua, nil
}

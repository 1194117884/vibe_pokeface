package model

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
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

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (s *UserDB) CreateUserWithAuth(ctx context.Context, user *User, auth *UserAuth) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, "INSERT INTO users (nickname, role) VALUES (?, ?)", user.Nickname, user.Role)
	if err != nil {
		if isDuplicateEntry(err) {
			return ErrDuplicateNickname
		}
		return err
	}
	user.ID, _ = result.LastInsertId()
	auth.UserID = user.ID

	_, err = tx.ExecContext(ctx,
		"INSERT INTO user_auths (user_id, provider, provider_uid, credential) VALUES (?, ?, ?, ?)",
		auth.UserID, auth.Provider, auth.ProviderUID, auth.Credential)
	if err != nil {
		if isDuplicateEntry(err) {
			return ErrDuplicateAuth
		}
		return err
	}

	return tx.Commit()
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

func (s *UserDB) ListUsers(ctx context.Context, offset, limit int) ([]User, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var users []User
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	return users, err
}

func (s *UserDB) SearchUsers(ctx context.Context, query string) ([]User, error) {
	var users []User
	like := "%" + query + "%"
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM users WHERE nickname LIKE ? ORDER BY created_at DESC LIMIT 50", like)
	return users, err
}

func (s *UserDB) UpdateUserStatus(ctx context.Context, userID int64, status int8) error {
	result, err := s.db.ExecContext(ctx, "UPDATE users SET status = ? WHERE id = ?", status, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *UserDB) GetUserCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users")
	return count, err
}

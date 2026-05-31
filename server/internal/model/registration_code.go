package model

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/jmoiron/sqlx"
)

// RegistrationCode represents a one-time-use registration code.
type RegistrationCode struct {
	ID            int64      `db:"id" json:"id"`
	Code          string     `db:"code" json:"code"`
	IsUsed        bool       `db:"is_used" json:"is_used"`
	IsDisabled    bool       `db:"is_disabled" json:"is_disabled"`
	UsedByUserID  *int64     `db:"used_by_user_id" json:"used_by_user_id,omitempty"`
	UsedAt        *time.Time `db:"used_at" json:"used_at,omitempty"`
	CreatedBy     int64      `db:"created_by" json:"created_by"`
	Note          string     `db:"note" json:"note"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// RegistrationCodeStore defines the interface for registration code operations.
type RegistrationCodeStore interface {
	Create(ctx context.Context, codes []*RegistrationCode) error
	FindByCode(ctx context.Context, code string) (*RegistrationCode, error)
	ClaimCode(ctx context.Context, code string, userID int64) (bool, error)
	List(ctx context.Context, offset, limit int, usedFilter *bool) ([]RegistrationCodeWithUser, int, error)
	Disable(ctx context.Context, id int64) error
	CountByStatus(ctx context.Context, used bool) (int, error)
}

// RegistrationCodeWithUser includes the user nickname for display.
type RegistrationCodeWithUser struct {
	RegistrationCode
	UsedByNickname *string `db:"used_by_nickname" json:"used_by_nickname,omitempty"`
}

// RegistrationCodeDB implements RegistrationCodeStore.
type RegistrationCodeDB struct {
	db *sqlx.DB
}

func NewRegistrationCodeDB(db *sqlx.DB) *RegistrationCodeDB {
	return &RegistrationCodeDB{db: db}
}

// GenerateCode creates a random 6-digit numeric string.
func GenerateCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// Create inserts multiple registration codes. Retries on code collision.
func (s *RegistrationCodeDB) Create(ctx context.Context, codes []*RegistrationCode) error {
	for _, c := range codes {
		for attempt := 0; attempt < 5; attempt++ {
			code, err := GenerateCode()
			if err != nil {
				return fmt.Errorf("generate code: %w", err)
			}
			c.Code = code

			_, err = s.db.ExecContext(ctx,
				`INSERT INTO registration_codes (code, created_by, note) VALUES (?, ?, ?)`,
				c.Code, c.CreatedBy, c.Note)
			if err == nil {
				break
			}
			if !isDuplicateEntry(err) {
				return err
			}
			if attempt == 4 {
				return fmt.Errorf("failed to generate unique code after 5 attempts")
			}
		}
	}
	return nil
}

// FindByCode looks up a registration code by its code string.
func (s *RegistrationCodeDB) FindByCode(ctx context.Context, code string) (*RegistrationCode, error) {
	var rc RegistrationCode
	err := s.db.GetContext(ctx, &rc, "SELECT * FROM registration_codes WHERE code = ?", code)
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

// ClaimCode atomically marks a code as used. Returns true if successful, false if already used/disabled.
func (s *RegistrationCodeDB) ClaimCode(ctx context.Context, code string, userID int64) (bool, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE registration_codes SET is_used = TRUE, used_by_user_id = ?, used_at = NOW() WHERE code = ? AND is_used = FALSE AND is_disabled = FALSE`,
		userID, code)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// List returns paginated registration codes, optionally filtered by used status.
func (s *RegistrationCodeDB) List(ctx context.Context, offset, limit int, usedFilter *bool) ([]RegistrationCodeWithUser, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	where := ""
	args := []interface{}{}
	if usedFilter != nil {
		where = " WHERE rc.is_used = ?"
		args = append(args, *usedFilter)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM registration_codes rc" + where
	if err := s.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := `SELECT rc.*, u.nickname AS used_by_nickname
		FROM registration_codes rc
		LEFT JOIN users u ON rc.used_by_user_id = u.id` + where +
		` ORDER BY rc.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	var codes []RegistrationCodeWithUser
	if err := s.db.SelectContext(ctx, &codes, query, args...); err != nil {
		return nil, 0, err
	}
	if codes == nil {
		codes = []RegistrationCodeWithUser{}
	}
	return codes, total, nil
}

// Disable marks an unused, non-disabled code as disabled. Returns sql.ErrNoRows if not found or already used.
func (s *RegistrationCodeDB) Disable(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE registration_codes SET is_disabled = TRUE WHERE id = ? AND is_used = FALSE AND is_disabled = FALSE`,
		id)
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

// CountByStatus returns the count of codes filtered by used status.
func (s *RegistrationCodeDB) CountByStatus(ctx context.Context, used bool) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM registration_codes WHERE is_used = ?", used)
	return count, err
}

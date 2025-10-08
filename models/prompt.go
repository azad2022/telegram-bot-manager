package models

import (
	"database/sql"
	"errors"
	"time"
)

type Prompt struct {
	ID         int
	UserID     int
	Content    string
	Response   string
	CreatedAt  time.Time
	IsFavorite bool
}

func CreatePrompt(db *sql.DB, userID int, content, response string) error {
	query := `
		INSERT INTO prompts (user_id, content, response, created_at)
		VALUES ($1, $2, $3, NOW());
	`
	_, err := db.Exec(query, userID, content, response)
	return err
}

func GetUserPrompts(db *sql.DB, userID int, limit int) ([]Prompt, error) {
	query := `
		SELECT id, user_id, content, response, created_at, false as is_favorite
		FROM prompts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2;
	`
	rows, err := db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []Prompt
	for rows.Next() {
		var p Prompt
		if err := rows.Scan(&p.ID, &p.UserID, &p.Content, &p.Response, &p.CreatedAt, &p.IsFavorite); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}

	return prompts, nil
}

// CheckPromptLimit بررسی می‌کند که آیا کاربر هنوز مجاز به ساخت پرامپت جدید هست یا نه.
func CheckPromptLimit(db *sql.DB, userID int, isVIP bool) (bool, int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM prompts WHERE user_id = $1;`, userID).Scan(&count)
	if err != nil {
		return false, 0, err
	}

	maxPrompts := 3
	if isVIP {
		maxPrompts = 10
	}

	remaining := maxPrompts - count
	if remaining <= 0 {
		return false, 0, nil
	}

	return true, remaining, nil
}

// DeleteOldPrompts حذف پرامپت‌های قدیمی در صورت تجاوز از محدودیت
func DeleteOldPrompts(db *sql.DB, userID int, isVIP bool) error {
	maxPrompts := 3
	if isVIP {
		maxPrompts = 10
	}

	query := `
		DELETE FROM prompts
		WHERE id IN (
			SELECT id FROM prompts
			WHERE user_id = $1
			ORDER BY created_at DESC
			OFFSET $2
		);
	`
	_, err := db.Exec(query, userID, maxPrompts)
	return err
}

// EnsurePromptTable ایجاد جدول در صورت نبود
func EnsurePromptTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS prompts (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			response TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			is_favorite BOOLEAN DEFAULT FALSE
		);
	`
	_, err := db.Exec(query)
	if err != nil {
		return errors.New("cannot create table prompts: " + err.Error())
	}
	return nil
}

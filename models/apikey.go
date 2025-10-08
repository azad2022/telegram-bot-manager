package models

import (
	"database/sql"
	"errors"
	"time"
)

// ساختار ذخیره کلید API
type APIKey struct {
	ID         int
	UserID     int64
	APIKey     string
	IsActive   bool
	CreatedAt  time.Time
}

// افزودن یا بروزرسانی کلید API
func SaveAPIKey(db *sql.DB, userID int64, key string) error {
	_, err := db.Exec(`
		INSERT INTO user_api_keys (user_id, api_key, is_active, created_at)
		VALUES ($1, $2, TRUE, NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET api_key = EXCLUDED.api_key, is_active = TRUE, created_at = NOW()
	`, userID, key)
	return err
}

// حذف کلید API کاربر
func DeleteAPIKey(db *sql.DB, userID int64) error {
	_, err := db.Exec(`DELETE FROM user_api_keys WHERE user_id = $1`, userID)
	return err
}

// دریافت کلید فعال کاربر
func GetActiveAPIKey(db *sql.DB, userID int64) (string, error) {
	var key string
	err := db.QueryRow(`SELECT api_key FROM user_api_keys WHERE user_id = $1 AND is_active = TRUE LIMIT 1`, userID).Scan(&key)
	if err == sql.ErrNoRows {
		return "", errors.New("کلید فعالی یافت نشد")
	}
	return key, err
}

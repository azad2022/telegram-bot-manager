package models

import (
	"database/sql"
	"errors"
	"time"

	"telegram-bot-manager/database"
)

// APIKey ساختار مدل کلید کاربران
type APIKey struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Key       string    `db:"api_key"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
}

// CreateTable اگر جدول وجود نداشت ایجاد می‌کند
func CreateTableAPIKeys() error {
	query := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		api_key TEXT NOT NULL,
		active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := database.DB.Exec(query)
	return err
}

// SaveAPIKey — ذخیره API جدید کاربر
func SaveAPIKey(userID int64, key string) error {
	if userID == 0 || key == "" {
		return errors.New("اطلاعات ورودی ناقص است")
	}

	// غیرفعال کردن کلیدهای قدیمی کاربر
	_, _ = database.DB.Exec("UPDATE api_keys SET active = FALSE WHERE user_id = $1", userID)

	// درج کلید جدید
	_, err := database.DB.Exec("INSERT INTO api_keys (user_id, api_key, active) VALUES ($1, $2, TRUE)", userID, key)
	return err
}

// GetActiveAPIKey — دریافت آخرین کلید فعال کاربر
func GetActiveAPIKey(userID int64) (string, error) {
	if userID == 0 {
		return "", errors.New("شناسه کاربر نامعتبر است")
	}

	var apiKey string
	err := database.DB.QueryRow("SELECT api_key FROM api_keys WHERE user_id=$1 AND active=TRUE ORDER BY id DESC LIMIT 1", userID).Scan(&apiKey)

	if err == sql.ErrNoRows {
		return "", nil
	}

	return apiKey, err
}

// DeleteAPIKey — حذف کلید کاربر
func DeleteAPIKey(userID int64) error {
	_, err := database.DB.Exec("DELETE FROM api_keys WHERE user_id=$1", userID)
	return err
}

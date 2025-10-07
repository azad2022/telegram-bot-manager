package models

import (
	"database/sql"
	"time"
)

type APIKey struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	APIKey    string    `json:"api_key"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenUsage struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Date      time.Time `json:"date"`
	TokensUsed int       `json:"tokens_used"`
	Cost      float64   `json:"cost"`
}

// ایجاد API Key جدید
func CreateAPIKey(db *sql.DB, userID int64, apiKey string) error {
	// غیرفعال کردن سایر API Keyهای کاربر
	err := DeactivateAllUserAPIKeys(db, userID)
	if err != nil {
		return err
	}

	// ایجاد API Key جدید به عنوان فعال
	query := `
		INSERT INTO api_keys (user_id, api_key, is_active, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = db.Exec(query, userID, apiKey, true, time.Now())
	return err
}

// دریافت API Key فعال کاربر
func GetActiveAPIKey(db *sql.DB, userID int64) (*APIKey, error) {
	query := `
		SELECT id, user_id, api_key, is_active, created_at
		FROM api_keys 
		WHERE user_id = $1 AND is_active = true
		LIMIT 1
	`
	
	apiKey := &APIKey{}
	err := db.QueryRow(query, userID).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.APIKey, &apiKey.IsActive, &apiKey.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // API Key فعالی وجود ندارد
		}
		return nil, err
	}
	
	return apiKey, nil
}

// دریافت تمام API Keyهای کاربر
func GetUserAPIKeys(db *sql.DB, userID int64) ([]APIKey, error) {
	query := `
		SELECT id, user_id, api_key, is_active, created_at
		FROM api_keys 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`
	
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var apiKeys []APIKey
	for rows.Next() {
		var apiKey APIKey
		err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.APIKey, &apiKey.IsActive, &apiKey.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, apiKey)
	}
	
	return apiKeys, nil
}

// غیرفعال کردن تمام API Keyهای کاربر
func DeactivateAllUserAPIKeys(db *sql.DB, userID int64) error {
	query := `UPDATE api_keys SET is_active = false WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	return err
}

// فعال کردن API Key خاص
func ActivateAPIKey(db *sql.DB, userID, apiKeyID int64) error {
	// غیرفعال کردن تمام API Keyهای کاربر
	err := DeactivateAllUserAPIKeys(db, userID)
	if err != nil {
		return err
	}

	// فعال کردن API Key انتخاب شده
	query := `UPDATE api_keys SET is_active = true WHERE id = $1 AND user_id = $2`
	_, err = db.Exec(query, apiKeyID, userID)
	return err
}

// حذف API Key
func DeleteAPIKey(db *sql.DB, apiKeyID int64) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := db.Exec(query, apiKeyID)
	return err
}

// ثبت مصرف توکن
func RecordTokenUsage(db *sql.DB, userID int64, tokensUsed int, cost float64) error {
	today := time.Now().Truncate(24 * time.Hour)
	
	query := `
		INSERT INTO token_usage (user_id, date, tokens_used, cost)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, date) 
		DO UPDATE SET tokens_used = token_usage.tokens_used + $3, cost = token_usage.cost + $4
	`
	_, err := db.Exec(query, userID, today, tokensUsed, cost)
	return err
}

// دریافت مصرف روزانه
func GetDailyUsage(db *sql.DB, userID int64, date time.Time) (*TokenUsage, error) {
	targetDate := date.Truncate(24 * time.Hour)
	
	query := `
		SELECT id, user_id, date, tokens_used, cost
		FROM token_usage 
		WHERE user_id = $1 AND date = $2
	`
	
	usage := &TokenUsage{}
	err := db.QueryRow(query, userID, targetDate).Scan(
		&usage.ID, &usage.UserID, &usage.Date, &usage.TokensUsed, &usage.Cost,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return &TokenUsage{
				UserID:    userID,
				Date:      targetDate,
				TokensUsed: 0,
				Cost:      0,
			}, nil
		}
		return nil, err
	}
	
	return usage, nil
}

// دریافت مصرف هفتگی
func GetWeeklyUsage(db *sql.DB, userID int64) (int, float64, error) {
	weekAgo := time.Now().AddDate(0, 0, -7)
	
	query := `
		SELECT COALESCE(SUM(tokens_used), 0), COALESCE(SUM(cost), 0)
		FROM token_usage 
		WHERE user_id = $1 AND date >= $2
	`
	
	var totalTokens int
	var totalCost float64
	err := db.QueryRow(query, userID, weekAgo).Scan(&totalTokens, &totalCost)
	if err != nil {
		return 0, 0, err
	}
	
	return totalTokens, totalCost, nil
}

// دریافت مصرف ماهانه
func GetMonthlyUsage(db *sql.DB, userID int64) (int, float64, error) {
	monthAgo := time.Now().AddDate(0, -1, 0)
	
	query := `
		SELECT COALESCE(SUM(tokens_used), 0), COALESCE(SUM(cost), 0)
		FROM token_usage 
		WHERE user_id = $1 AND date >= $2
	`
	
	var totalTokens int
	var totalCost float64
	err := db.QueryRow(query, userID, monthAgo).Scan(&totalTokens, &totalCost)
	if err != nil {
		return 0, 0, err
	}
	
	return totalTokens, totalCost, nil
}

// بررسی سقف مصرف
func CheckUsageLimit(db *sql.DB, userID int64, isVIP bool) (bool, int, error) {
	dailyUsage, err := GetDailyUsage(db, userID, time.Now())
	if err != nil {
		return false, 0, err
	}

	var dailyLimit int
	if isVIP {
		dailyLimit = 100000 // کاربران VIP: 100,000 توکن در روز
	} else {
		dailyLimit = 10000 // کاربران عادی: 10,000 توکن در روز
	}

	remaining := dailyLimit - dailyUsage.TokensUsed
	isWithinLimit := remaining > 0

	return isWithinLimit, remaining, nil
}

// دریافت آمار مصرف کلی
func GetUsageStats(db *sql.DB, userID int64) (daily, weekly, monthly int, err error) {
	// مصرف روزانه
	dailyUsage, err := GetDailyUsage(db, userID, time.Now())
	if err != nil {
		return 0, 0, 0, err
	}
	daily = dailyUsage.TokensUsed

	// مصرف هفتگی
	weekly, _, err = GetWeeklyUsage(db, userID)
	if err != nil {
		return 0, 0, 0, err
	}

	// مصرف ماهانه
	monthly, _, err = GetMonthlyUsage(db, userID)
	if err != nil {
		return 0, 0, 0, err
	}

	return daily, weekly, monthly, nil
}

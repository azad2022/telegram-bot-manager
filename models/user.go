package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID          int64        `json:"id"`
	TelegramID  int64        `json:"telegram_id"`
	Username    string       `json:"username"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	Phone       string       `json:"phone"`
	IsVIP       bool         `json:"is_vip"`
	VIPUntil    sql.NullTime `json:"vip_until"`
	InviteCount int          `json:"invite_count"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ایجاد کاربر جدید
func CreateUser(db *sql.DB, telegramID int64, username, firstName, lastName string) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (telegram_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			updated_at = EXCLUDED.updated_at
	`
	_, err := db.Exec(query, telegramID, username, firstName, lastName, time.Now(), time.Now())
	return err
}

// دریافت کاربر بر اساس آیدی تلگرام
func GetUserByTelegramID(db *sql.DB, telegramID int64) (*User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, phone, 
		       is_vip, vip_until, invite_count, created_at, updated_at
		FROM users 
		WHERE telegram_id = $1
	`
	
	user := &User{}
	var vipUntil sql.NullTime
	
	err := db.QueryRow(query, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName,
		&user.Phone, &user.IsVIP, &vipUntil, &user.InviteCount, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // کاربر وجود ندارد
		}
		return nil, err
	}
	
	user.VIPUntil = vipUntil
	return user, nil
}

// آپدیت شماره تلفن کاربر
func UpdateUserPhone(db *sql.DB, telegramID int64, phone string) error {
	query := `UPDATE users SET phone = $1, updated_at = $2 WHERE telegram_id = $3`
	_, err := db.Exec(query, phone, time.Now(), telegramID)
	return err
}

// فعال‌سازی VIP برای کاربر
func ActivateVIP(db *sql.DB, telegramID int64, durationDays int) error {
	vipUntil := time.Now().AddDate(0, 0, durationDays)
	
	query := `
		UPDATE users 
		SET is_vip = true, vip_until = $1, updated_at = $2 
		WHERE telegram_id = $3
	`
	_, err := db.Exec(query, vipUntil, time.Now(), telegramID)
	return err
}

// غیرفعال کردن VIP
func DeactivateVIP(db *sql.DB, telegramID int64) error {
	query := `
		UPDATE users 
		SET is_vip = false, vip_until = NULL, updated_at = $1 
		WHERE telegram_id = $2
	`
	_, err := db.Exec(query, time.Now(), telegramID)
	return err
}

// افزایش تعداد دعوت‌های موفق
func IncrementInviteCount(db *sql.DB, telegramID int64) error {
	query := `
		UPDATE users 
		SET invite_count = invite_count + 1, updated_at = $1 
		WHERE telegram_id = $2
	`
	_, err := db.Exec(query, time.Now(), telegramID)
	return err
}

// بررسی انقضای VIP کاربران
func CheckVIPExpiration(db *sql.DB) error {
	query := `
		UPDATE users 
		SET is_vip = false, vip_until = NULL, updated_at = $1 
		WHERE is_vip = true AND vip_until < $2
	`
	_, err := db.Exec(query, time.Now(), time.Now())
	return err
}

// دریافت کاربران VIP
func GetVIPUsers(db *sql.DB) ([]User, error) {
	query := `
		SELECT telegram_id, username, first_name, vip_until 
		FROM users 
		WHERE is_vip = true AND vip_until > $1
		ORDER BY vip_until DESC
	`
	
	rows, err := db.Query(query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []User
	for rows.Next() {
		var user User
		var vipUntil sql.NullTime
		
		err := rows.Scan(&user.TelegramID, &user.Username, &user.FirstName, &vipUntil)
		if err != nil {
			return nil, err
		}
		
		user.VIPUntil = vipUntil
		user.IsVIP = true
		users = append(users, user)
	}
	
	return users, nil
}

// دریافت آمار کاربران
func GetUserStats(db *sql.DB) (totalUsers, vipUsers int, err error) {
	// تعداد کل کاربران
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return 0, 0, err
	}
	
	// تعداد کاربران VIP
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE is_vip = true AND vip_until > $1", time.Now()).Scan(&vipUsers)
	if err != nil {
		return 0, 0, err
	}
	
	return totalUsers, vipUsers, nil
}

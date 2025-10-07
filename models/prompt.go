package models

import (
	"database/sql"
	"time"
)

type Prompt struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// ایجاد پرامپت جدید
func CreatePrompt(db *sql.DB, userID int64, title, content string) error {
	// غیرفعال کردن سایر پرامپت‌های کاربر
	err := DeactivateAllUserPrompts(db, userID)
	if err != nil {
		return err
	}

	// ایجاد پرامپت جدید به عنوان فعال
	query := `
		INSERT INTO prompts (user_id, title, content, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = db.Exec(query, userID, title, content, true, time.Now())
	return err
}

// دریافت تمام پرامپت‌های یک کاربر
func GetUserPrompts(db *sql.DB, userID int64) ([]Prompt, error) {
	query := `
		SELECT id, user_id, title, content, is_active, created_at
		FROM prompts 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`
	
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var prompts []Prompt
	for rows.Next() {
		var prompt Prompt
		err := rows.Scan(
			&prompt.ID, &prompt.UserID, &prompt.Title, &prompt.Content, 
			&prompt.IsActive, &prompt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		prompts = append(prompts, prompt)
	}
	
	return prompts, nil
}

// دریافت پرامپت فعال کاربر
func GetActivePrompt(db *sql.DB, userID int64) (*Prompt, error) {
	query := `
		SELECT id, user_id, title, content, is_active, created_at
		FROM prompts 
		WHERE user_id = $1 AND is_active = true
		LIMIT 1
	`
	
	prompt := &Prompt{}
	err := db.QueryRow(query, userID).Scan(
		&prompt.ID, &prompt.UserID, &prompt.Title, &prompt.Content, 
		&prompt.IsActive, &prompt.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // پرامپت فعالی وجود ندارد
		}
		return nil, err
	}
	
	return prompt, nil
}

// انتخاب پرامپت فعال
func SetActivePrompt(db *sql.DB, userID, promptID int64) error {
	// غیرفعال کردن تمام پرامپت‌های کاربر
	err := DeactivateAllUserPrompts(db, userID)
	if err != nil {
		return err
	}

	// فعال کردن پرامپت انتخاب شده
	query := `UPDATE prompts SET is_active = true WHERE id = $1 AND user_id = $2`
	_, err = db.Exec(query, promptID, userID)
	return err
}

// غیرفعال کردن تمام پرامپت‌های کاربر
func DeactivateAllUserPrompts(db *sql.DB, userID int64) error {
	query := `UPDATE prompts SET is_active = false WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	return err
}

// به‌روزرسانی پرامپت
func UpdatePrompt(db *sql.DB, promptID int64, title, content string) error {
	query := `UPDATE prompts SET title = $1, content = $2 WHERE id = $3`
	_, err := db.Exec(query, title, content, promptID)
	return err
}

// حذف پرامپت
func DeletePrompt(db *sql.DB, promptID int64) error {
	query := `DELETE FROM prompts WHERE id = $1`
	_, err := db.Exec(query, promptID)
	return err
}

// بررسی تعداد پرامپت‌های کاربر
func GetUserPromptCount(db *sql.DB, userID int64, isVIP bool) (int, error) {
	var maxPrompts int
	if isVIP {
		maxPrompts = 10 // کاربران VIP تا 10 پرامپت
	} else {
		maxPrompts = 3 // کاربران عادی تا 3 پرامپت
	}

	query := `SELECT COUNT(*) FROM prompts WHERE user_id = $1`
	var count int
	err := db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// بررسی امکان ایجاد پرامپت جدید
func CanCreatePrompt(db *sql.DB, userID int64, isVIP bool) (bool, int, error) {
	count, err := GetUserPromptCount(db, userID, isVIP)
	if err != nil {
		return false, 0, err
	}

	var maxPrompts int
	if isVIP {
		maxPrompts = 10
	} else {
		maxPrompts = 3
	}

	return count < maxPrompts, maxPrompts - count, nil
}

// دریافت اطلاعات پرامپت بر اساس ID
func GetPromptByID(db *sql.DB, promptID int64) (*Prompt, error) {
	query := `
		SELECT id, user_id, title, content, is_active, created_at
		FROM prompts 
		WHERE id = $1
	`
	
	prompt := &Prompt{}
	err := db.QueryRow(query, promptID).Scan(
		&prompt.ID, &prompt.UserID, &prompt.Title, &prompt.Content, 
		&prompt.IsActive, &prompt.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // پرامپت وجود ندارد
		}
		return nil, err
	}
	
	return prompt, nil
}

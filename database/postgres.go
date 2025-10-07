package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitPostgreSQL(connectionString string) error {
	var err error
	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("خطا در اتصال به PostgreSQL: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("خطا در ping PostgreSQL: %v", err)
	}

	log.Println("✅ اتصال به PostgreSQL برقرار شد")
	return createTables()
}

func createTables() error {
	// جدول کاربران
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			telegram_id BIGINT UNIQUE NOT NULL,
			username VARCHAR(255),
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			phone VARCHAR(20),
			is_vip BOOLEAN DEFAULT FALSE,
			vip_until TIMESTAMP,
			invite_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// جدول پرامپت‌ها
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS prompts (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			is_active BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(telegram_id)
		)
	`)
	if err != nil {
		return err
	}

	// جدول API Keys
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS api_keys (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			api_key TEXT NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(telegram_id)
		)
	`)
	if err != nil {
		return err
	}

	log.Println("✅ جداول دیتابیس ایجاد شدند")
	return nil
}

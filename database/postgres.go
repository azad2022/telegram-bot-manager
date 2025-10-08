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

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("خطا در ping PostgreSQL: %v", err)
	}

	log.Println("✅ اتصال به PostgreSQL برقرار شد")
	
	if err := createTables(); err != nil {
		return fmt.Errorf("خطا در ایجاد جداول: %v", err)
	}

	log.Println("✅ جداول دیتابیس با موفقیت ایجاد شدند")
	return nil
}

func createTables() error {
	if err := createUsersTable(); err != nil {
		return err
	}
	if err := createPromptsTable(); err != nil {
		return err
	}
	if err := createAPIKeysTable(); err != nil {
		return err
	}
	if err := createTokenUsageTable(); err != nil {
		return err
	}
	if err := createChannelsTable(); err != nil {
		return err
	}
	if err := createGroupsTable(); err != nil {
		return err
	}
	if err := createPaymentRequestsTable(); err != nil {
		return err
	}
	if err := createPaymentLinksTable(); err != nil {
		return err
	}
	if err := createReferralsTable(); err != nil {
		return err
	}
	if err := createSystemLogsTable(); err != nil {
		return err
	}

	return nil
}

func createUsersTable() error {
	query := `
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
	);

	CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
	CREATE INDEX IF NOT EXISTS idx_users_is_vip ON users(is_vip);
	CREATE INDEX IF NOT EXISTS idx_users_vip_until ON users(vip_until);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول users: %v", err)
	}
	log.Println("✓ جدول users ایجاد شد")
	return nil
}

func createPromptsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS prompts (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		is_active BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_prompts_user_id ON prompts(user_id);
	CREATE INDEX IF NOT EXISTS idx_prompts_is_active ON prompts(is_active);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول prompts: %v", err)
	}
	log.Println("✓ جدول prompts ایجاد شد")
	return nil
}

func createAPIKeysTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		api_key TEXT NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
	CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول api_keys: %v", err)
	}
	log.Println("✓ جدول api_keys ایجاد شد")
	return nil
}

func createTokenUsageTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS token_usage (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		date DATE NOT NULL,
		tokens_used INTEGER DEFAULT 0,
		cost DECIMAL(10, 4) DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, date),
		FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_token_usage_user_date ON token_usage(user_id, date);
	CREATE INDEX IF NOT EXISTS idx_token_usage_date ON token_usage(date);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول token_usage: %v", err)
	}
	log.Println("✓ جدول token_usage ایجاد شد")
	return nil
}

func createChannelsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS channels (
		id SERIAL PRIMARY KEY,
		owner_id BIGINT NOT NULL,
		channel_id VARCHAR(255) UNIQUE NOT NULL,
		channel_title VARCHAR(255),
		prompt TEXT,
		schedule_time VARCHAR(5),
		posts_per_batch INTEGER DEFAULT 1,
		is_active BOOLEAN DEFAULT FALSE,
		last_post_at TIMESTAMP,
		total_posts INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (owner_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
		CONSTRAINT check_schedule_time CHECK (schedule_time ~ '^([0-1][0-9]|2[0-3]):[0-5][0-9]$'),
		CONSTRAINT check_posts_per_batch CHECK (posts_per_batch BETWEEN 1 AND 10)
	);

	CREATE INDEX IF NOT EXISTS idx_channels_owner_id ON channels(owner_id);
	CREATE INDEX IF NOT EXISTS idx_channels_is_active ON channels(is_active);
	CREATE INDEX IF NOT EXISTS idx_channels_schedule_time ON channels(schedule_time);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول channels: %v", err)
	}
	log.Println("✓ جدول channels ایجاد شد")
	return nil
}

func createGroupsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS groups (
		id SERIAL PRIMARY KEY,
		group_id BIGINT UNIQUE NOT NULL,
		group_title VARCHAR(255),
		owner_id BIGINT NOT NULL,
		footer_text TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		rate_limit INTEGER DEFAULT 5,
		total_questions INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (owner_id) REFERENCES users(telegram_id) ON DELETE SET NULL,
		CONSTRAINT check_rate_limit CHECK (rate_limit BETWEEN 1 AND 100)
	);

	CREATE INDEX IF NOT EXISTS idx_groups_group_id ON groups(group_id);
	CREATE INDEX IF NOT EXISTS idx_groups_owner_id ON groups(owner_id);
	CREATE INDEX IF NOT EXISTS idx_groups_is_active ON groups(is_active);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول groups: %v", err)
	}
	log.Println("✓ جدول groups ایجاد شد")
	return nil
}

func createPaymentRequestsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS payment_requests (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		plan VARCHAR(50) NOT NULL,
		amount DECIMAL(10, 2),
		status VARCHAR(20) DEFAULT 'pending',
		payment_proof TEXT,
		rejection_reason TEXT,
		processed_by BIGINT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
		CONSTRAINT check_status CHECK (status IN ('pending', 'approved', 'rejected')),
		CONSTRAINT check_plan CHECK (plan IN ('1month', '3months', '6months', '1year'))
	);

	CREATE INDEX IF NOT EXISTS idx_payment_requests_user_id ON payment_requests(user_id);
	CREATE INDEX IF NOT EXISTS idx_payment_requests_status ON payment_requests(status);
	CREATE INDEX IF NOT EXISTS idx_payment_requests_created_at ON payment_requests(created_at DESC);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول payment_requests: %v", err)
	}
	log.Println("✓ جدول payment_requests ایجاد شد")
	return nil
}

func createPaymentLinksTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS payment_links (
		id SERIAL PRIMARY KEY,
		plan VARCHAR(50) UNIQUE NOT NULL,
		link TEXT NOT NULL,
		price DECIMAL(10, 2),
		duration_days INTEGER NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT check_plan CHECK (plan IN ('1month', '3months', '6months', '1year'))
	);

	INSERT INTO payment_links (plan, link, price, duration_days, description)
	VALUES 
		('1month', 'https://example.com/pay/1month', 50000.00, 30, 'اشتراک یک ماهه'),
		('3months', 'https://example.com/pay/3months', 140000.00, 90, 'اشتراک سه ماهه'),
		('6months', 'https://example.com/pay/6months', 260000.00, 180, 'اشتراک شش ماهه'),
		('1year', 'https://example.com/pay/1year', 480000.00, 365, 'اشتراک یک ساله')
	ON CONFLICT (plan) DO NOTHING;
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول payment_links: %v", err)
	}
	log.Println("✓ جدول payment_links ایجاد شد")
	return nil
}

func createReferralsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS referrals (
		id SERIAL PRIMARY KEY,
		referrer_id BIGINT NOT NULL,
		referred_id BIGINT NOT NULL,
		reward_claimed BOOLEAN DEFAULT FALSE,
		reward_type VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(referrer_id, referred_id),
		FOREIGN KEY (referrer_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
		FOREIGN KEY (referred_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
		CONSTRAINT check_different_users CHECK (referrer_id != referred_id)
	);

	CREATE INDEX IF NOT EXISTS idx_referrals_referrer_id ON referrals(referrer_id);
	CREATE INDEX IF NOT EXISTS idx_referrals_referred_id ON referrals(referred_id);
	CREATE INDEX IF NOT EXISTS idx_referrals_reward_claimed ON referrals(reward_claimed);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول referrals: %v", err)
	}
	log.Println("✓ جدول referrals ایجاد شد")
	return nil
}

func createSystemLogsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS system_logs (
		id SERIAL PRIMARY KEY,
		log_type VARCHAR(50) NOT NULL,
		user_id BIGINT,
		action VARCHAR(255),
		details TEXT,
		ip_address VARCHAR(45),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(telegram_id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_system_logs_log_type ON system_logs(log_type);
	CREATE INDEX IF NOT EXISTS idx_system_logs_user_id ON system_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_system_logs_created_at ON system_logs(created_at DESC);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("خطا در ایجاد جدول system_logs: %v", err)
	}
	log.Println("✓ جدول system_logs ایجاد شد")
	return nil
}

func DropAllTables() error {
	tables := []string{
		"system_logs",
		"referrals",
		"payment_requests",
		"payment_links",
		"groups",
		"channels",
		"token_usage",
		"api_keys",
		"prompts",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("خطا در حذف جدول %s: %v", table, err)
		}
		log.Printf("🗑️ جدول %s حذف شد", table)
	}

	return nil
}

func GetDatabaseStats() (map[string]int, error) {
	stats := make(map[string]int)
	tables := []string{
		"users", "prompts", "api_keys", "token_usage",
		"channels", "groups", "payment_requests", "referrals",
	}

	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := DB.QueryRow(query).Scan(&count)
		if err != nil {
			return nil, err
		}
		stats[table] = count
	}

	return stats, nil
}

func CheckHealth() error {
	if DB == nil {
		return fmt.Errorf("دیتابیس اتصال ندارد")
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("خطا در ping دیتابیس: %v", err)
	}

	return nil
}

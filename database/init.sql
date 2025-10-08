-- ================================
-- 🧠 Telegram Bot Manager - Database Schema
-- ================================

-- 👤 جدول کاربران
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    vip BOOLEAN DEFAULT FALSE,
    referral_code VARCHAR(64),
    invited_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 🔑 جدول کلیدهای API کاربران (OpenAI API Keys)
CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    api_key TEXT NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    usage_count INT DEFAULT 0,
    last_used TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 💬 جدول پرامپت‌های سفارشی کاربران
CREATE TABLE IF NOT EXISTS prompts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    prompt TEXT NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 📊 جدول مصرف توکن‌ها (Token Usage Logs)
CREATE TABLE IF NOT EXISTS token_usage (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    api_key_id INT REFERENCES api_keys(id) ON DELETE CASCADE,
    tokens_used INT NOT NULL,
    model VARCHAR(100),
    cost_usd DECIMAL(10, 5),
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 📢 جدول کانال‌ها (برای Scheduler)
CREATE TABLE IF NOT EXISTS channels (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    channel_id BIGINT UNIQUE NOT NULL,
    prompt_id INT REFERENCES prompts(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    schedule_interval INT DEFAULT 60, -- دقیقه
    last_post TIMESTAMP,
    next_post TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 🧮 ایندکس‌های پرکاربرد
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users (telegram_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys (user_id);
CREATE INDEX IF NOT EXISTS idx_prompts_user_id ON prompts (user_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_user_id ON token_usage (user_id);
CREATE INDEX IF NOT EXISTS idx_channels_user_id ON channels (user_id);

-- ✅ پایان

package main

type Config struct {
	TelegramToken string
	AdminID       int64
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
}

func LoadConfig() *Config {
	return &Config{
		TelegramToken: "8407008563:AAHBQpjUh60bHqpxOfAJqEfTmicNO6IfEl0",
		AdminID:       269758292,
		DatabaseURL:   "postgres://bot_user:bot_password_123@localhost:5433/telegram_bot?sslmode=disable",
		RedisURL:      "localhost:6380",
		RedisPassword: "redis_password_123",
	}
}

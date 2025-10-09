package main

import (
	"os"
)

type Config struct {
	BotToken    string
	DatabaseURL string
}

func LoadConfig() Config {
	return Config{
		BotToken: os.Getenv("BOT_TOKEN"),

		// ✅ کانکشن درست به دیتابیس اصلی با پورت استاندارد
		DatabaseURL: "postgres://bot_user:bot_password_123@localhost:5432/telegram_bot_manager?sslmode=disable",
	}
}

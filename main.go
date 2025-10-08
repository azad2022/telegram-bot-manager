package main

import (
	"log"
	"os"
	"time"

	"gopkg.in/telebot.v3"

	"telegram-bot-manager/database"
	"telegram-bot-manager/handlers"
)

func main() {
	log.Println("🚀 در حال راه‌اندازی ربات...")

	// 1️⃣ اتصال به Postgres
	if err := database.ConnectPostgres(); err != nil {
		log.Fatalf("❌ خطا در اتصال به Postgres: %v", err)
	}

	// 2️⃣ اتصال به Redis
	if err := database.ConnectRedis(); err != nil {
		log.Fatalf("❌ خطا در اتصال به Redis: %v", err)
	}

	// 3️⃣ تنظیم توکن ربات
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ متغیر BOT_TOKEN تنظیم نشده است.")
	}

	// 4️⃣ پیکربندی ربات
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("❌ خطا در ساخت ربات: %v", err)
	}

	// 5️⃣ ثبت هندلرهای اصلی
	handlers.HandlePrivateMessage(bot)
	// در آینده: handlers.HandleGroupMessage(bot)
	// در آینده: handlers.HandleAdmin(bot)

	// 6️⃣ شروع به کار
	log.Println("🤖 ربات با موفقیت راه‌اندازی شد و در حال اجراست...")
	bot.Start()
}

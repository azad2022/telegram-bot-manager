package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"telegram-bot-manager/database"
	"telegram-bot-manager/handlers"
	"telegram-bot-manager/services"

	"gopkg.in/telebot.v3"
)

func main() {
	// بارگذاری تنظیمات
	config := LoadConfig()

	// اتصال به PostgreSQL
	err := database.InitPostgreSQL(config.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ خطا در اتصال به PostgreSQL: %v", err)
	}
	log.Println("✅ اتصال به PostgreSQL برقرار شد")

	// اتصال به Redis
	err = database.InitRedis(config.RedisURL, config.RedisPassword)
	if err != nil {
		log.Fatalf("❌ خطا در اتصال به Redis: %v", err)
	}
	log.Println("✅ اتصال به Redis برقرار شد")

	// تنظیمات ربات تلگرام
	pref := telebot.Settings{
		Token:  config.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("❌ خطا در راه‌اندازی ربات: %v", err)
	}

	log.Println("🤖 ربات با موفقیت راه‌اندازی شد")

	// هندلرهای اصلی
	handlers.HandlePrivateMessage(bot, database.DB)
	handlers.HandleGroupMessages(bot, database.DB)

	// سیستم زمان‌بندی تولید محتوا
	scheduler := services.NewScheduler(bot, database.DB)
	go scheduler.Start()
	go scheduler.StartMaintenance()

	// مدیریت سیگنال‌های خروج
	go waitForShutdown(bot)

	// شروع ربات
	log.Println("🚀 ربات در حال اجراست...")
	bot.Start()
}

// تابع کنترل خاموشی امن
func waitForShutdown(bot *telebot.Bot) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 دریافت سیگنال خاموشی... در حال بستن منابع")

	bot.Stop()

	if database.DB != nil {
		_ = database.DB.Close()
		log.Println("✅ اتصال PostgreSQL بسته شد")
	}
	if database.RDB != nil {
		_ = database.RDB.Close()
		log.Println("✅ اتصال Redis بسته شد")
	}

	log.Println("✅ ربات با موفقیت متوقف شد")
	os.Exit(0)
}

cat > main.go << 'EOF'
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"telegram-bot-manager/database"
	"telegram-bot-manager/handlers"
	"telegram-bot-manager/models"
	"telegram-bot-manager/services"

	"gopkg.in/telebot.v3"
)

func main() {
	// بارگذاری تنظیمات
	config := LoadConfig()

	log.Println("🚀 شروع راه‌اندازی ربات...")

	// اتصال به دیتابیس PostgreSQL
	err := database.InitPostgreSQL(config.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ خطا در اتصال به PostgreSQL: %v", err)
	}

	// اتصال به Redis
	err = database.InitRedis(config.RedisURL, config.RedisPassword)
	if err != nil {
		log.Fatalf("❌ خطا در اتصال به Redis: %v", err)
	}

	// تنظیمات ربات تلگرام
	pref := telebot.Settings{
		Token:  config.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("✅ اتصال به دیتابیس‌ها برقرار شد")
	log.Println("🤖 ربات راه‌اندازی شد...")

	// راه‌اندازی هندلرها
	setupHandlers(bot)

	// راه‌اندازی scheduler
	scheduler := services.NewScheduler(bot, database.DB)
	go scheduler.Start()
	go scheduler.StartMaintenance()

	// شروع ربات
	go bot.Start()

	log.Println("🎯 ربات آماده دریافت پیام‌ها...")

	// مدیریت خاموشی گران‌قدر
	waitForShutdown()
}

func setupHandlers(bot *telebot.Bot) {
	// هندلر چت خصوصی
	bot.Handle("/start", func(c telebot.Context) error {
		return handlers.HandleStartCommand(c, database.DB)
	})

	// هندلر مدیریت منوهای اصلی
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		// اگر در چت خصوصی هستیم
		if c.Chat().Type == telebot.ChatPrivate {
			return handlers.HandlePrivateText(c, database.DB)
		}
		
		// اگر در گروه هستیم و پیام با * شروع شده
		text := c.Text()
		if len(text) > 0 && text[0] == '*' {
			return handlers.HandleGroupQuestion(c, database.DB, text)
		}
		
		return nil
	})

	// هندلر برای دکمه‌های منو
	setupMenuHandlers(bot)
}

func setupMenuHandlers(bot *telebot.Bot) {
	// منوی مدیریت پرامپت‌ها
	bot.Handle("🧠 مدیریت پرامپت‌ها", func(c telebot.Context) error {
		return handlers.HandlePromptManagement(c, database.DB)
	})

	// منوی مدیریت API
	bot.Handle("🔑 مدیریت API", func(c telebot.Context) error {
		return handlers.HandleAPIManagement(c, database.DB)
	})

	// منوی مشاهده مصرف
	bot.Handle("📊 مشاهده مصرف", func(c telebot.Context) error {
		return handlers.HandleUsageStats(c, database.DB)
	})

	// منوی تنظیمات مدل
	bot.Handle("⚙️ تنظیمات مدل", func(c telebot.Context) error {
		return handlers.HandleModelSettings(c, database.DB)
	})

	// منوی تنظیمات کانال
	bot.Handle("🔧 تنظیمات کانال", func(c telebot.Context) error {
		return handlers.HandleChannelSettings(c, database.DB)
	})

	// منوی تنظیمات گروه
	bot.Handle("🔨 تنظیمات گروه", func(c telebot.Context) error {
		return handlers.HandleGroupSettings(c, database.DB)
	})

	// منوی امتیازگیری
	bot.Handle("🎯 امتیازگیری", func(c telebot.Context) error {
		return handlers.HandleInvitationSystem(c, database.DB)
	})

	// منوی راهنمای ربات
	bot.Handle("📣 راهنمای ربات", func(c telebot.Context) error {
		return handlers.HandleHelpGuide(c)
	})

	// پنل مدیریت (فقط برای سازنده)
	bot.Handle("🛠️ پنل مدیریت", func(c telebot.Context) error {
		if c.Sender().ID != 269758292 {
			return c.Send("⛔ دسترسی denied")
		}
		return handlers.HandleAdminPanel(c, database.DB)
	})
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Println("🛑 دریافت سیگنال خاموشی...")
	
	// بستن اتصال به دیتابیس
	if database.DB != nil {
		database.DB.Close()
	}
	
	if database.RDB != nil {
		database.RDB.Close()
	}
	
	log.Println("✅ ربات با موفقیت متوقف شد")
	os.Exit(0)
}
EOF

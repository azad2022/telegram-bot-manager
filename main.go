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

	// ۱️⃣ اتصال به PostgreSQL
	db, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("❌  خطا در اتصال به PostgreSQL: %v", err)
	}
	log.Println("✅  اتصال موفق به PostgreSQL برقرار شد.")

	// ۲️⃣ اتصال به Redis
	rdb, err := database.ConnectRedis()
	if err != nil {
		log.Fatalf("❌  خطا در اتصال به Redis: %v", err)
	}
	log.Println("✅  اتصال موفق به Redis برقرار شد.")

	// جلوگیری از خطای استفاده‌نشده
	_ = db
	_ = rdb

	// ۳️⃣ خواندن توکن ربات از متغیر محیطی
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("❌  متغیر BOT_TOKEN تنظیم نشده است.")
	}

	// ۴️⃣ پیکربندی ربات
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	// ۵️⃣ ساخت نمونه‌ی ربات
	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("❌  خطا در ایجاد ربات: %v", err)
	}

	// ۶️⃣ تعریف هندلرهای اصلی
	bot.Handle("/start", func(c telebot.Context) error {
		msg := "سلام 👋\nمن آماده‌ام — از دکمه‌ها یا ارسال پیام استفاده کن.\n\nدکمه‌ها:\n➕ /addapi - افزودن API\n🗑️ /removeapi - حذف API\n(پس از افزودن API، هر پیام شما به ChatGPT ارسال می‌شود.)"
		return c.Send(msg)
	})

	// ⚙️ هندلرهای مدیریت API
	bot.Handle("/addapi", handlers.HandleAddAPI(bot, db))
	bot.Handle("/removeapi", handlers.HandleRemoveAPI(bot, db))

	// ✅ شروع کار ربات
	log.Println("🤖 ربات با موفقیت راه‌اندازی شد و در حال اجراست...")
	bot.Start()
}

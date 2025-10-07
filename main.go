package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/telebot.v3"
)

func main() {
	// تنظیمات ربات
	pref := telebot.Settings{
		Token:  "8407008563:AAHBQpjUh60bHqpxOfAJqEfTmicNO6IfEl0",
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// هندلرهای اصلی
	bot.Handle("/start", func(c telebot.Context) error {
		menu := &telebot.ReplyMarkup{}
		menu.Reply(
			menu.Row(menu.Text("🧠 مدیریت پرامپت‌ها")),
			menu.Row(menu.Text("🔑 مدیریت API")),
			menu.Row(menu.Text("📊 مشاهده مصرف")),
			menu.Row(menu.Text("⚙️ تنظیمات مدل")),
			menu.Row(menu.Text("🔧 تنظیمات کانال VIP")),
			menu.Row(menu.Text("🔨 تنظیمات گروه VIP")),
			menu.Row(menu.Text("🎯 امتیازگیری")),
			menu.Row(menu.Text("📣 راهنمای ربات")),
		)
		
		if c.Sender().ID == 269758292 {
			menu.Reply(menu.Row(menu.Text("🛠️ پنل مدیریت")))
		}

		return c.Send("🤖 به ربات ChatGPT خوش آمدید!", menu)
	})

	// هندلرهای منو
	bot.Handle("🧠 مدیریت پرامپت‌ها", func(c telebot.Context) error {
		// کد مدیریت پرامپت‌ها
		return c.Send("مدیریت پرامپت‌ها")
	})

	bot.Handle("🛠️ پنل مدیریت", func(c telebot.Context) error {
		if c.Sender().ID != 269758292 {
			return c.Send("⛔ دسترسی denied")
		}
		
		// کد پنل مدیریت
		return c.Send("پنل مدیریت")
	})

	log.Println("ربات شروع به کار کرد...")
	bot.Start()

	// انتظار برای سیگنال خاتمه
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("ربات متوقف شد...")
}

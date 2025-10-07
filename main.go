package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/telebot.v3"
)

func main() {
	pref := telebot.Settings{
		Token:  "8407008563:AAHBQpjUh60bHqpxOfAJqEfTmicNO6IfEl0",
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// منوی اصلی برای همه کاربران
	bot.Handle("/start", func(c telebot.Context) error {
		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		
		// ایجاد ردیف‌های منو
		row1 := menu.Row(menu.Text("🧠 مدیریت پرامپت‌ها"), menu.Text("🔑 مدیریت API"))
		row2 := menu.Row(menu.Text("📊 مشاهده مصرف"), menu.Text("⚙️ تنظیمات مدل"))
		row3 := menu.Row(menu.Text("🔧 تنظیمات کانال"), menu.Text("🔨 تنظیمات گروه"))
		row4 := menu.Row(menu.Text("🎯 امتیازگیری"), menu.Text("📣 راهنمای ربات"))
		
		rows := []telebot.Row{row1, row2, row3, row4}
		
		// اگر سازنده هست، پنل مدیریت اضافه کن
		if c.Sender().ID == 269758292 {
			row5 := menu.Row(menu.Text("🛠️ پنل مدیریت"))
			rows = append(rows, row5)
		}
		
		menu.Reply(rows...)
		
		return c.Send("🤖 به ربات ChatGPT خوش آمدید!\n\n" +
			"از منوی زیر انتخاب کنید:", menu)
	})

	// هندلرهای منو
	bot.Handle("🧠 مدیریت پرامپت‌ها", func(c telebot.Context) error {
		return c.Send("🔄 به زودی: مدیریت پرامپت‌ها\n\n" +
			"در این بخش می‌توانید:\n" +
			"• پرامپت‌های شخصی اضافه کنید\n" +
			"• پرامپت فعال انتخاب کنید\n" +
			"• پرامپت‌ها را ویرایش یا حذف کنید")
	})

	bot.Handle("🔑 مدیریت API", func(c telebot.Context) error {
		return c.Send("🔑 مدیریت API Keys\n\n" +
			"در این بخش می‌توانید:\n" +
			"• API Key خود را اضافه کنید\n" +
			"• مصرف توکن را مشاهده کنید\n" +
			"• هشدار مصرف دریافت کنید")
	})

	bot.Handle("📊 مشاهده مصرف", func(c telebot.Context) error {
		return c.Send("📊 مشاهده مصرف\n\n" +
			"مصرف امروز: 0 توکن\n" +
			"مصرف این ماه: 0 توکن\n" +
			"سقف مصرف: نامحدود")
	})

	bot.Handle("⚙️ تنظیمات مدل", func(c telebot.Context) error {
		return c.Send("🤖 تنظیمات مدل\n\n" +
			"مدل فعلی: GPT-3.5 Turbo\n\n" +
			"گزینه‌های موجود:\n" +
			"• GPT-3.5 Turbo\n" +
			"• GPT-4\n" +
			"• GPT-4 Turbo")
	})

	bot.Handle("🔧 تنظیمات کانال", func(c telebot.Context) error {
		return c.Send("📢 تنظیمات کانال\n\n" +
			"ویژگی‌های VIP:\n" +
			"• تولید محتوای خودکار\n" +
			"• زمان‌بندی انتشار\n" +
			"• مدیریت پرامپت کانال")
	})

	bot.Handle("🔨 تنظیمات گروه", func(c telebot.Context) error {
		return c.Send("💬 تنظیمات گروه\n\n" +
			"ویژگی‌های VIP:\n" +
			"• پاسخ‌گویی خودکار\n" +
			"• تنظیم پرامپت گروه\n" +
			"• مدیریت محدودیت‌ها")
	})

	bot.Handle("🎯 امتیازگیری", func(c telebot.Context) error {
		return c.Send("🎯 سیستم امتیازگیری\n\n" +
			"با دعوت دوستان امتیاز بگیرید:\n" +
			"• هر 20 دعوت = 1 روز VIP\n" +
			"• لینک دعوت اختصاصی\n" +
			"• پیگیری تعداد دعوت‌ها")
	})

	bot.Handle("📣 راهنمای ربات", func(c telebot.Context) error {
		return c.Send("📣 راهنمای ربات\n\n" +
			"نحوه استفاده:\n" +
			"• در گروه: *سوال خود را بنویسید\n" +
			"• در PV: از منو استفاده کنید\n" +
			"• در کانال: تولید محتوای خودکار")
	})

	bot.Handle("🛠️ پنل مدیریت", func(c telebot.Context) error {
		if c.Sender().ID != 269758292 {
			return c.Send("⛔ دسترسی denied")
		}
		
		return c.Send("🛠️ پنل مدیریت سازنده\n\n" +
			"آمار سیستم:\n" +
			"• کاربران: 1\n" +
			"• گروه‌ها: 0\n" +
			"• کانال‌ها: 0\n\n" +
			"مدیریت:\n" +
			"• مشاهده کاربران\n" +
			"• مدیریت VIP\n" +
			"• تنظیمات پرداخت")
	})

	log.Println("🤖 ربات شروع به کار کرد...")
	bot.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("🛑 ربات متوقف شد...")
}

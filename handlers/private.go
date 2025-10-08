package handlers

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/telebot.v3"

	"telegram-bot-manager/database"
	"telegram-bot-manager/models"
	"telegram-bot-manager/services"
)

// HandlePrivateMessage ثبت هندلرهای پیوی (private chat)
func HandlePrivateMessage(bot *telebot.Bot) {
	ctx := context.Background()

	// اطمینان از وجود جدول api_keys (بدون ضرر اگر قبلاً ساخته شده)
	_ = models.CreateTableAPIKeys()

	// /start - منوی ساده
	bot.Handle("/start", func(c telebot.Context) error {
		menu := "سلام 👋\n\nمن آماده‌ام — از دکمه‌ها یا ارسال پیام استفاده کن.\n\nدکمه‌ها:\n➕ افزودن API\n🗑️ حذف API\n(پس از افزودن API، هر پیام شما به ChatGPT ارسال می‌شود.)"
		return c.Send(menu)
	})

	// دستور افزودن API (هم دکمه یا متن)
	bot.Handle("/addapikey", func(c telebot.Context) error {
		uid := c.Sender().ID
		key := fmt.Sprintf("state:%d", uid)
		_ = database.RDB.Set(ctx, key, "awaiting_api", 5*time.Minute).Err()
		return c.Send("لطفاً کلید API خود را ارسال کنید (در یک پیام).")
	})

	// دستور حذف API
	bot.Handle("/delapikey", func(c telebot.Context) error {
		uid := c.Sender().ID
		if err := models.DeleteAPIKey(uid); err != nil {
			return c.Send("خطا در حذف کلید: " + err.Error())
		}
		return c.Send("✅ همهٔ کلیدهای شما حذف شدند.")
	})

	// هندل متن‌ها — شامل دریافت API و ارسال پیام به OpenAI
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		uid := c.Sender().ID
		text := c.Text()
		stateKey := fmt.Sprintf("state:%d", uid)

		// 1) اگر کاربر در حالت ارسال API باشد
		if st, _ := database.RDB.Get(ctx, stateKey).Result(); st == "awaiting_api" {
			// ذخیره API
			if err := models.SaveAPIKey(uid, text); err != nil {
				return c.Send("❌ خطا در ذخیره‌سازی کلید: " + err.Error())
			}
			database.RDB.Del(ctx, stateKey)
			return c.Send("✅ کلید شما با موفقیت ذخیره شد. حالا هر پیامی بفرستید از API شما استفاده خواهد شد.")
		}

		// 2) اگر پیام شامل دستور افزودن یا حذف به صورت متن ساده باشه
		if text == "➕ افزودن API" || text == "/addapikey" {
			_ = database.RDB.Set(ctx, stateKey, "awaiting_api", 5*time.Minute).Err()
			return c.Send("لطفاً حالا کلید API خود را ارسال کنید.")
		}
		if text == "🗑️ حذف API" || text == "/delapikey" {
			if err := models.DeleteAPIKey(uid); err != nil {
				return c.Send("خطا در حذف کلید: " + err.Error())
			}
			return c.Send("✅ کلیدهای شما حذف شدند.")
		}

		// 3) حالت عادی: باید یک API فعال برای کاربر وجود داشته باشد
		apiKey, err := models.GetActiveAPIKey(uid)
		if err != nil {
			return c.Send("❌ خطا در خواندن کلید از دیتابیس.")
		}
		if apiKey == "" {
			return c.Send("⚠️ شما هنوز API ثبت نکرده‌اید. برای ثبت: /addapikey یا '➕ افزودن API'")
		}

		// 4) اعمال rate-limit ساده: 5 درخواست در دقیقه
		rlKey := fmt.Sprintf("rl:%d", uid)
		count, _ := database.RDB.Incr(ctx, rlKey).Result()
		if count == 1 {
			database.RDB.Expire(ctx, rlKey, time.Minute)
		}
		if count > 5 {
			return c.Send("🚫 محدودیت: بیش از حد مجاز پیام فرستادید. لطفاً بعد از چند ثانیه دوباره تلاش کنید.")
		}

		// 5) ارسال به OpenAI با apiKey کاربر
		// مدل پیش‌فرض: gpt-3.5-turbo (در آینده می‌تونی از انتخاب کاربر استفاده کنی)
		model := "gpt-3.5-turbo"
		resp, err := services.SendChatWithKey(apiKey, model, text)
		if err != nil {
			// لاگ خطا در سرور (اختیاری) و پیام مناسب به کاربر
			// fmt.Printf("openai error: %v\n", err)
			return c.Send("❌ خطا در تماس با OpenAI: " + err.Error())
		}

		// 6) ارسال پاسخ به کاربر
		_, sendErr := c.Bot().Send(c.Sender(), resp)
		if sendErr != nil {
			return c.Send("❌ خطا در ارسال پاسخ: " + sendErr.Error())
		}
		return nil
	})
}

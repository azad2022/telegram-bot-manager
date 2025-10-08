package handlers

import (
	"database/sql"
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"
	"telegram-bot-manager/models"
)

// HandlePrivateMessage - مدیریت پیام‌های خصوصی کاربران
func HandlePrivateMessage(bot *telebot.Bot, db *sql.DB) {
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		user := c.Sender()
		userID := user.ID

		// ثبت خودکار کاربر در دیتابیس در صورت عدم وجود
		_ = models.CreateUserIfNotExists(db, userID, user.Username, user.FirstName, user.LastName)

		text := strings.TrimSpace(c.Text())

		switch text {
		case "/start":
			msg := "سلام 👋\nمن ربات مدیریت ChatGPT هستم.\n\n" +
				"می‌تونی کلید API خودت رو اضافه یا حذف کنی:\n\n" +
				"➕ افزودن API: فقط کلیدت رو بفرست (مثلاً sk-...)\n" +
				"🗑️ حذف API: دستور /removeapi رو بفرست."
			return c.Send(msg)

		case "/removeapi":
			err := models.DeleteAPIKey(db, userID)
			if err != nil {
				return c.Send("❌ خطا در حذف API از دیتابیس.")
			}
			return c.Send("✅ کلید API شما با موفقیت حذف شد.")

		default:
			// اگر متن شامل "sk-" بود، یعنی کلید API جدید ارسال شده
			if len(text) > 10 && strings.HasPrefix(text, "sk-") {
				err := models.SaveAPIKey(db, userID, text)
				if err != nil {
					return c.Send(fmt.Sprintf("❌ خطا در ذخیره کلید در دیتابیس: %v", err))
				}
				return c.Send("✅ کلید API با موفقیت ذخیره شد.")
			}

			// بررسی وجود API Key فعال
			apiKey, err := models.GetActiveAPIKey(db, userID)
			if err != nil {
				return c.Send("❌ خطا در خواندن کلید از دیتابیس.")
			}
			if apiKey == "" {
				return c.Send("⚠️ هنوز کلید API ثبت نکرده‌اید.\nکلید خود را ارسال کنید تا ثبت شود.")
			}

			// پاسخ موقت (در آینده جایگزین ChatGPT API می‌شود)
			return c.Send(fmt.Sprintf("📩 پیام شما دریافت شد و آماده ارسال به ChatGPT با کلید:\n%s", apiKey))
		}
	})
}

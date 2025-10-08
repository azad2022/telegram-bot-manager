package handlers

import (
	"database/sql"
	"fmt"
	"gopkg.in/telebot.v3"
	"strings"
	"telegram-bot-manager/models"
)

// -----------------------------
// مدیریت API کاربران
// -----------------------------

// HandleAddAPI - افزودن کلید API جدید
func HandleAddAPI(bot *telebot.Bot, db *sql.DB) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := c.Sender().ID

		args := strings.TrimSpace(c.Message().Payload)
		if args == "" {
			return c.Send("🔑 لطفاً کلید API خود را بعد از دستور ارسال کنید.\nمثال:\n`/addapi sk-abc123`", &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
		}

		if !strings.HasPrefix(args, "sk-") {
			return c.Send("❌ فرمت کلید API معتبر نیست. باید با `sk-` شروع شود.")
		}

		err := models.SaveAPIKey(db, userID, args)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ خطا در ذخیره کلید: %v", err))
		}

		return c.Send("✅ کلید API شما با موفقیت ذخیره شد.")
	}
}

// HandleRemoveAPI - حذف کلید API کاربر
func HandleRemoveAPI(bot *telebot.Bot, db *sql.DB) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := c.Sender().ID

		err := models.DeleteAPIKey(db, userID)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ خطا در حذف کلید: %v", err))
		}

		return c.Send("🗑️ کلید API شما حذف شد.")
	}
}

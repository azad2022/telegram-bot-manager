package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"telegram-bot-manager/database"
	"telegram-bot-manager/models"
)

// HandleAdminPanel - مدیریت پنل ادمین
func HandleAdminPanel(c telebot.Context, db *sql.DB) error {
	// بررسی دسترسی
	if c.Sender().ID != 269758292 {
		return c.Send("⛔ دسترسی denied")
	}

	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	// دکمه‌های پنل مدیریت
	btnStats := menu.Text("📊 آمار کامل")
	btnSearch := menu.Text("🔍 جستجوی کاربر")
	btnVIP := menu.Text("👑 مدیریت VIP")
	btnPayments := menu.Text("💳 درخواست‌های پرداخت")
	btnLinks := menu.Text("🔗 تنظیم لینک‌ها")
	btnInvites := menu.Text("📋 گزارش دعوت‌ها")
	btnBack := menu.Text("🔙 بازگشت")

	menu.Reply(
		menu.Row(btnStats),
		menu.Row(btnSearch, btnVIP),
		menu.Row(btnPayments, btnLinks),
		menu.Row(btnInvites),
		menu.Row(btnBack),
	)

	// هندلرهای پنل مدیریت
	bot.Handle("📊 آمار کامل", func(c telebot.Context) error {
		return handleAdminStats(c, db)
	})

	bot.Handle("🔍 جستجوی کاربر", func(c telebot.Context) error {
		return handleUserSearch(c, db)
	})

	bot.Handle("👑 مدیریت VIP", func(c telebot.Context) error {
		return handleVIPManagement(c, db)
	})

	bot.Handle("💳 درخواست‌های پرداخت", func(c telebot.Context) error {
		return handlePaymentRequests(c, db)
	})

	bot.Handle("🔗 تنظیم لینک‌ها", func(c telebot.Context) error {
		return handlePaymentLinks(c, db)
	})

	bot.Handle("📋 گزارش دعوت‌ها", func(c telebot.Context) error {
		return handleInvitationReports(c, db)
	})

	return c.Send("🛠️ پنل مدیریت سازنده\n\nاز گزینه‌های زیر انتخاب کنید:", menu)
}

// آمار کامل سیستم
func handleAdminStats(c telebot.Context, db *sql.DB) error {
	// آمار کاربران
	totalUsers, vipUsers, err := models.GetUserStats(db)
	if err != nil {
		return c.Send("❌ خطا در دریافت آمار کاربران")
	}

	// آمار مصرف امروز
	var dailyUsage int
	err = db.QueryRow(`
		SELECT COALESCE(SUM(tokens_used), 0) 
		FROM token_usage 
		WHERE date = CURRENT_DATE
	`).Scan(&dailyUsage)
	if err != nil {
		dailyUsage = 0
	}

	// آمار گروه‌ها و کانال‌ها
	var groupCount, channelCount int
	// این بخش بعداً با جدول گروه‌ها و کانال‌ها تکمیل می‌شود

	message := fmt.Sprintf(
		"📊 آمار کامل سیستم\n\n"+
			"👥 کاربران:\n"+
			"• کل کاربران: %d\n"+
			"• کاربران VIP: %d\n"+
			"• کاربران عادی: %d\n\n"+
			"📈 مصرف امروز:\n"+
			"• توکن مصرف شده: %d\n"+
			"• هزینه تقریبی: %.2f تومان\n\n"+
			"💬 محیط‌ها:\n"+
			"• گروه‌ها: %d\n"+
			"• کانال‌ها: %d",
		totalUsers, vipUsers, totalUsers-vipUsers,
		dailyUsage, float64(dailyUsage)*0.002*30000, // تقریب هزینه
		groupCount, channelCount,
	)

	return c.Send(message)
}

// جستجوی کاربر
func handleUserSearch(c telebot.Context, db *sql.DB) error {
	return c.Send("لطفاً آیدی عددی کاربر را وارد کنید:")
}

// مدیریت VIP
func handleVIPManagement(c telebot.Context, db *sql.DB) error {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnAddVIP := menu.Text("⭐ افزودن VIP")
	btnRemoveVIP := menu.Text("🗑️ حذف VIP")
	btnListVIP := menu.Text("📋 لیست کاربران VIP")
	btnBack := menu.Text("🔙 بازگشت")

	menu.Reply(
		menu.Row(btnAddVIP, btnRemoveVIP),
		menu.Row(btnListVIP),
		menu.Row(btnBack),
	)

	// هندلرهای مدیریت VIP
	bot.Handle("⭐ افزودن VIP", func(c telebot.Context) error {
		return handleAddVIP(c, db)
	})

	bot.Handle("🗑️ حذف VIP", func(c telebot.Context) error {
		return handleRemoveVIP(c, db)
	})

	bot.Handle("📋 لیست کاربران VIP", func(c telebot.Context) error {
		return handleListVIPUsers(c, db)
	})

	return c.Send("👑 مدیریت کاربران VIP\n\nاز گزینه‌های زیر انتخاب کنید:", menu)
}

// افزودن کاربر به VIP
func handleAddVIP(c telebot.Context, db *sql.DB) error {
	return c.Send("لطفاً آیدی کاربر و مدت VIP (به روز) را وارد کنید:\n\nفرمت: آیدی مدت\nمثال: 123456789 30")
}

// حذف کاربر از VIP
func handleRemoveVIP(c telebot.Context, db *sql.DB) error {
	return c.Send("لطفاً آیدی کاربری که می‌خواهید از VIP حذف کنید را وارد کنید:")
}

// لیست کاربران VIP
func handleListVIPUsers(c telebot.Context, db *sql.DB) error {
	vipUsers, err := models.GetVIPUsers(db)
	if err != nil {
		return c.Send("❌ خطا در دریافت لیست کاربران VIP")
	}

	if len(vipUsers) == 0 {
		return c.Send("📭 هیچ کاربر VIP‌ای وجود ندارد")
	}

	var message strings.Builder
	message.WriteString("👑 لیست کاربران VIP:\n\n")

	for i, user := range vipUsers {
		username := "بدون یوزرنیم"
		if user.Username != "" {
			username = "@" + user.Username
		}

		message.WriteString(fmt.Sprintf(
			"%d. %s (%s)\n",
			i+1, username, user.FirstName,
		))
	}

	return c.Send(message.String())
}

// مدیریت درخواست‌های پرداخت
func handlePaymentRequests(c telebot.Context, db *sql.DB) error {
	// این بخش بعداً با جدول پرداخت‌ها تکمیل می‌شود
	return c.Send("💳 سیستم درخواست‌های پرداخت\n\nبه زودی فعال خواهد شد...")
}

// تنظیم لینک‌های پرداخت
func handlePaymentLinks(c telebot.Context, db *sql.DB) error {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btn1Month := menu.Text("⭐ ۱ ماه")
	btn3Months := menu.Text("⭐⭐ ۳ ماه")
	btn6Months := menu.Text("⭐⭐⭐ ۶ ماه")
	btn1Year := menu.Text("💎 ۱ سال")
	btnBack := menu.Text("🔙 بازگشت")

	menu.Reply(
		menu.Row(btn1Month, btn3Months),
		menu.Row(btn6Months, btn1Year),
		menu.Row(btnBack),
	)

	// هندلرهای تنظیم لینک پرداخت
	bot.Handle("⭐ ۱ ماه", func(c telebot.Context) error {
		return handleSetPaymentLink(c, "1month")
	})

	bot.Handle("⭐⭐ ۳ ماه", func(c telebot.Context) error {
		return handleSetPaymentLink(c, "3months")
	})

	bot.Handle("⭐⭐⭐ ۶ ماه", func(c telebot.Context) error {
		return handleSetPaymentLink(c, "6months")
	})

	bot.Handle("💎 ۱ سال", func(c telebot.Context) error {
		return handleSetPaymentLink(c, "1year")
	})

	return c.Send("🔗 تنظیم لینک‌های پرداخت\n\nلینک پرداخت کدام پلن را می‌خواهید تنظیم کنید؟", menu)
}

func handleSetPaymentLink(c telebot.Context, plan string) error {
	planNames := map[string]string{
		"1month":  "۱ ماه",
		"3months": "۳ ماه",
		"6months": "۶ ماه",
		"1year":   "۱ سال",
	}

	return c.Send(fmt.Sprintf(
		"لطفاً لینک پرداخت برای پلن «%s» را وارد کنید:",
		planNames[plan],
	))
}

// گزارش دعوت‌ها
func handleInvitationReports(c telebot.Context, db *sql.DB) error {
	// دریافت کاربران براساس تعداد دعوت
	rows, err := db.Query(`
		SELECT telegram_id, username, first_name, invite_count 
		FROM users 
		WHERE invite_count > 0 
		ORDER BY invite_count DESC 
		LIMIT 20
	`)
	if err != nil {
		return c.Send("❌ خطا در دریافت گزارش دعوت‌ها")
	}
	defer rows.Close()

	var message strings.Builder
	message.WriteString("📋 گزارش دعوت‌های موفق\n\n")

	count := 0
	for rows.Next() {
		var userID int64
		var username, firstName string
		var inviteCount int

		err := rows.Scan(&userID, &username, &firstName, &inviteCount)
		if err != nil {
			continue
		}

		count++
		if username == "" {
			username = "بدون یوزرنیم"
		} else {
			username = "@" + username
		}

		message.WriteString(fmt.Sprintf(
			"%d. %s - %d دعوت\n",
			count, username, inviteCount,
		))
	}

	if count == 0 {
		message.WriteString("📭 هیچ دعوت موفقی ثبت نشده است")
	}

	return c.Send(message.String())
}

// پردازش دستورات متنی در پنل مدیریت
func HandleAdminText(c telebot.Context, db *sql.DB) error {
	text := c.Text()

	// اگر عدد است، احتمالاً آیدی کاربر است
	if _, err := strconv.ParseInt(text, 10, 64); err == nil {
		return handleUserInfo(c, db, text)
	}

	// اگر شامل فاصله است، احتمالاً افزودن VIP است
	if strings.Contains(text, " ") {
		parts := strings.Split(text, " ")
		if len(parts) == 2 {
			if _, err := strconv.Atoi(parts[1]); err == nil {
				return processAddVIP(c, db, parts[0], parts[1])
			}
		}
	}

	return c.Send("❌ دستور نامعتبر. لطفاً دوباره تلاش کنید.")
}

// نمایش اطلاعات کاربر
func handleUserInfo(c telebot.Context, db *sql.DB, userIDStr string) error {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return c.Send("❌ آیدی کاربر نامعتبر است")
	}

	user, err := models.GetUserByTelegramID(db, userID)
	if err != nil || user == nil {
		return c.Send("❌ کاربر یافت نشد")
	}

	vipStatus := "عادی"
	if user.IsVIP {
		vipStatus = "VIP"
		if user.VIPUntil.Valid {
			vipStatus += fmt.Sprintf(" (تا %s)", user.VIPUntil.Time.Format("2006-01-02"))
		}
	}

	message := fmt.Sprintf(
		"👤 اطلاعات کاربر\n\n"+
			"🔸 آیدی: %d\n"+
			"🔸 نام: %s %s\n"+
			"🔸 یوزرنیم: %s\n"+
			"🔸 وضعیت: %s\n"+
			"🔸 تعداد دعوت: %d\n"+
			"🔸 تاریخ عضویت: %s",
		user.TelegramID,
		user.FirstName, user.LastName,
		getUsername(user.Username),
		vipStatus,
		user.InviteCount,
		user.CreatedAt.Format("2006-01-02"),
	)

	menu := &telebot.ReplyMarkup{}
	if user.IsVIP {
		btnRemove := menu.Data("🗑️ حذف VIP", "remove_vip", userIDStr)
		menu.Inline(menu.Row(btnRemove))
	} else {
		btnAdd1 := menu.Data("⭐ ۱ ماه", "add_vip", userIDStr+"_30")
		btnAdd3 := menu.Data("⭐⭐ ۳ ماه", "add_vip", userIDStr+"_90")
		menu.Inline(menu.Row(btnAdd1, btnAdd3))
	}

	return c.Send(message, menu)
}

// افزودن VIP به کاربر
func processAddVIP(c telebot.Context, db *sql.DB, userIDStr, daysStr string) error {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return c.Send("❌ آیدی کاربر نامعتبر است")
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return c.Send("❌ تعداد روز نامعتبر است")
	}

	err = models.ActivateVIP(db, userID, days)
	if err != nil {
		return c.Send("❌ خطا در فعال‌سازی VIP")
	}

	return c.Send(fmt.Sprintf(
		"✅ کاربر با آیدی %d به مدت %d روز به VIP ارتقا یافت",
		userID, days,
	))
}

// تابع کمکی برای نمایش یوزرنیم
func getUsername(username string) string {
	if username == "" {
		return "بدون یوزرنیم"
	}
	return "@" + username
}

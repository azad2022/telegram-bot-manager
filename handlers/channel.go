package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telebot.v3"

	"telegram-bot-manager/database"
	"telegram-bot-manager/models"
	"telegram-bot-manager/services"
)

// ChannelConfig - تنظیمات کانال
type ChannelConfig struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	ChannelID    string    `json:"channel_id"`
	ChannelTitle string    `json:"channel_title"`
	Prompt       string    `json:"prompt"`
	ScheduleTime string    `json:"schedule_time"`
	PostsPerBatch int      `json:"posts_per_batch"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// HandleChannelSettings - مدیریت تنظیمات کانال
func HandleChannelSettings(c telebot.Context, db *sql.DB) error {
	userID := c.Sender().ID
	
	// بررسی VIP بودن کاربر
	user, err := models.GetUserByTelegramID(db, userID)
	if err != nil || user == nil || !user.IsVIP {
		menu := &telebot.ReplyMarkup{}
		btnVIP := menu.URL("🎯 ارتقاء به VIP", "https://t.me/gpt_yourbot?start=vip_request")
		menu.Inline(menu.Row(btnVIP))
		
		return c.Send(
			"⛔ این قابلیت مخصوص کاربران VIP است\n\n"+
				"با ارتقاء به VIP می‌توانید:\n"+
				"• تولید محتوای خودکار در کانال\n"+
				"• زمان‌بندی انتشار پست‌ها\n"+
				"• تنظیم پرامپت اختصاصی کانال",
			menu,
		)
	}

	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	// دریافت تنظیمات کانال کاربر
	channelConfig, err := getChannelConfig(db, userID)
	if err != nil {
		log.Printf("خطا در دریافت تنظیمات کانال: %v", err)
	}

	// دکمه‌های منوی کانال
	btnSetChannel := menu.Text("📢 تنظیم آیدی کانال")
	btnSetPrompt := menu.Text("📝 تنظیم پرامپت")
	btnSetSchedule := menu.Text("⏰ تنظیم زمان انتشار")
	btnSetPosts := menu.Text("🔢 تنظیم تعداد پست")
	btnToggle := menu.Text("🔄 فعال/غیرفعال")
	btnStatus := menu.Text("📊 وضعیت کانال")
	btnBack := menu.Text("🔙 بازگشت")

	menu.Reply(
		menu.Row(btnSetChannel, btnSetPrompt),
		menu.Row(btnSetSchedule, btnSetPosts),
		menu.Row(btnToggle, btnStatus),
		menu.Row(btnBack),
	)

	// هندلرهای منوی کانال
	bot.Handle("📢 تنظیم آیدی کانال", func(c telebot.Context) error {
		return handleSetChannelID(c, db, userID)
	})

	bot.Handle("📝 تنظیم پرامپت", func(c telebot.Context) error {
		return handleSetChannelPrompt(c, db, userID)
	})

	bot.Handle("⏰ تنظیم زمان انتشار", func(c telebot.Context) error {
		return handleSetScheduleTime(c, db, userID)
	})

	bot.Handle("🔢 تنظیم تعداد پست", func(c telebot.Context) error {
		return handleSetPostsPerBatch(c, db, userID)
	})

	bot.Handle("🔄 فعال/غیرفعال", func(c telebot.Context) error {
		return handleToggleChannel(c, db, userID, channelConfig)
	})

	bot.Handle("📊 وضعیت کانال", func(c telebot.Context) error {
		return handleChannelStatus(c, db, userID, channelConfig)
	})

	// پیام خوش‌آمد
	message := "📢 مدیریت کانال VIP\n\n"
	if channelConfig != nil {
		status := "🔴 غیرفعال"
		if channelConfig.IsActive {
			status = "🟢 فعال"
		}
		
		message += fmt.Sprintf(
			"کانال: %s\n"+
				"وضعیت: %s\n"+
				"زمان انتشار: %s\n"+
				"تعداد پست: %d\n\n",
			channelConfig.ChannelTitle, status,
			channelConfig.ScheduleTime, channelConfig.PostsPerBatch,
		)
	} else {
		message += "هنوز کانالی تنظیم نکرده‌اید.\n\n"
	}
	
	message += "از گزینه‌های زیر انتخاب کنید:"

	return c.Send(message, menu)
}

// تنظیم آیدی کانال
func handleSetChannelID(c telebot.Context, db *sql.DB, userID int64) error {
	return c.Send("لطفاً آیدی کانال خود را وارد کنید:\n\n" +
		"فرمت: @channel_username\n" +
		"یا: https://t.me/channel_username\n\n" +
		"⚠️注意: ابتدا ربات را در کانال ادمین کنید")
}

// تنظیم پرامپت کانال
func handleSetChannelPrompt(c telebot.Context, db *sql.DB, userID int64) error {
	return c.Send("لطفاً پرامپت مخصوص کانال خود را وارد کنید:\n\n" +
		"مثال:\n" +
		"«تو یک تولیدکننده محتوای آموزشی هستی. روزانه یک نکته آموزشی در مورد برنامه‌نویسی تولید کن. محتوا باید کاربردی و قابل فهم باشد.»")
}

// تنظیم زمان انتشار
func handleSetScheduleTime(c telebot.Context, db *sql.DB, userID int64) error {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btn9AM := menu.Text("⏰ ۹:۰۰ صبح")
	btn12PM := menu.Text("⏰ ۱۲:۰۰ ظهر")
	btn6PM := menu.Text("⏰ ۱۸:۰۰ عصر")
	btn9PM := menu.Text("⏰ ۲۱:۰۰ شب")
	btnCustom := menu.Text("⏰ زمان دلخواه")
	btnBack := menu.Text("🔙 بازگشت")

	menu.Reply(
		menu.Row(btn9AM, btn12PM),
		menu.Row(btn6PM, btn9PM),
		menu.Row(btnCustom),
		menu.Row(btnBack),
	)

	// هندلرهای زمان‌بندی
	bot.Handle("⏰ ۹:۰۰ صبح", func(c telebot.Context) error {
		return saveScheduleTime(c, db, userID, "09:00")
	})

	bot.Handle("⏰ ۱۲:۰۰ ظهر", func(c telebot.Context) error {
		return saveScheduleTime(c, db, userID, "12:00")
	})

	bot.Handle("⏰ ۱۸:۰۰ عصر", func(c telebot.Context) error {
		return saveScheduleTime(c, db, userID, "18:00")
	})

	bot.Handle("⏰ ۲۱:۰۰ شب", func(c telebot.Context) error {
		return saveScheduleTime(c, db, userID, "21:00")
	})

	bot.Handle("⏰ زمان دلخواه", func(c telebot.Context) error {
		return c.Send("لطفاً زمان مورد نظر را به فرمت HH:MM وارد کنید:\n\nمثال: 08:30 یا 14:45")
	})

	return c.Send("⏰ تنظیم زمان انتشار\n\nزمان مورد نظر برای انتشار خودکار پست‌ها را انتخاب کنید:", menu)
}

// تنظیم تعداد پست
func handleSetPostsPerBatch(c telebot.Context, db *sql.DB, userID int64) error {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btn1 := menu.Text("1️⃣ ۱ پست")
	btn2 := menu.Text("2️⃣ ۲ پست")
	btn3 := menu.Text("3️⃣ ۳ پست")
	btn5 := menu.Text("5️⃣ ۵ پست")
	btnBack := menu.Text("🔙 بازگشت")

	menu.Reply(
		menu.Row(btn1, btn2),
		menu.Row(btn3, btn5),
		menu.Row(btnBack),
	)

	// هندلرهای تعداد پست
	bot.Handle("1️⃣ ۱ پست", func(c telebot.Context) error {
		return savePostsPerBatch(c, db, userID, 1)
	})

	bot.Handle("2️⃣ ۲ پست", func(c telebot.Context) error {
		return savePostsPerBatch(c, db, userID, 2)
	})

	bot.Handle("3️⃣ ۳ پست", func(c telebot.Context) error {
		return savePostsPerBatch(c, db, userID, 3)
	})

	bot.Handle("5️⃣ ۵ پست", func(c telebot.Context) error {
		return savePostsPerBatch(c, db, userID, 5)
	})

	return c.Send("🔢 تنظیم تعداد پست\n\nتعداد پست‌هایی که در هر نوبت منتشر شوند را انتخاب کنید:", menu)
}

// فعال/غیرفعال کردن کانال
func handleToggleChannel(c telebot.Context, db *sql.DB, userID int64, config *ChannelConfig) error {
	if config == nil {
		return c.Send("❌ ابتدا باید کانال خود را تنظیم کنید.")
	}

	// بررسی ادمین بودن ربات در کانال
	isAdmin, err := checkBotAdminStatus(c.Bot(), config.ChannelID)
	if err != nil {
		return c.Send("❌ خطا در بررسی وضعیت ربات در کانال")
	}

	if !isAdmin {
		return c.Send("❌ ربات در کانال ادمین نیست. لطفاً ابتدا ربات را ادمین کنید.")
	}

	newStatus := !config.IsActive
	err = updateChannelStatus(db, userID, newStatus)
	if err != nil {
		return c.Send("❌ خطا در تغییر وضعیت کانال")
	}

	statusText := "غیرفعال"
	if newStatus {
		statusText = "فعال"
	}

	return c.Send(fmt.Sprintf("✅ وضعیت کانال به «%s» تغییر یافت.", statusText))
}

// نمایش وضعیت کانال
func handleChannelStatus(c telebot.Context, db *sql.DB, userID int64, config *ChannelConfig) error {
	if config == nil {
		return c.Send("❌ هنوز کانالی تنظیم نکرده‌اید.")
	}

	// بررسی ادمین بودن ربات
	isAdmin, err := checkBotAdminStatus(c.Bot(), config.ChannelID)
	if err != nil {
		log.Printf("خطا در بررسی وضعیت ادمین: %v", err)
	}

	adminStatus := "❌ نیست"
	if isAdmin {
		adminStatus = "✅ هست"
	}

	channelStatus := "🔴 غیرفعال"
	if config.IsActive {
		channelStatus = "🟢 فعال"
	}

	message := fmt.Sprintf(
		"📊 وضعیت کانال\n\n"+
			"📢 کانال: %s\n"+
			"🔸 وضعیت: %s\n"+
			"🔸 ربات ادمین: %s\n"+
			"⏰ زمان انتشار: %s\n"+
			"🔢 تعداد پست: %d\n"+
			"📝 طول پرامپت: %d کاراکتر\n\n",
		config.ChannelTitle, channelStatus, adminStatus,
		config.ScheduleTime, config.PostsPerBatch, len(config.Prompt),
	)

	if !isAdmin {
		message += "⚠️ برای فعال‌سازی، ربات را در کانال ادمین کنید."
	}

	return c.Send(message)
}

// ذخیره زمان انتشار
func saveScheduleTime(c telebot.Context, db *sql.DB, userID int64, scheduleTime string) error {
	err := updateChannelSchedule(db, userID, scheduleTime)
	if err != nil {
		return c.Send("❌ خطا در ذخیره زمان انتشار")
	}

	return c.Send(fmt.Sprintf("✅ زمان انتشار به «%s» تنظیم شد.", scheduleTime))
}

// ذخیره تعداد پست
func savePostsPerBatch(c telebot.Context, db *sql.DB, userID int64, posts int) error {
	err := updateChannelPosts(db, userID, posts)
	if err != nil {
		return c.Send("❌ خطا در ذخیره تعداد پست")
	}

	return c.Send(fmt.Sprintf("✅ تعداد پست به «%d» تنظیم شد.", posts))
}

// پردازش پیام‌های متنی برای تنظیمات کانال
func HandleChannelText(c telebot.Context, db *sql.DB) error {
	text := c.Text()
	userID := c.Sender().ID

	// بررسی اینکه کاربر در حال تنظیم کانال است
	if strings.HasPrefix(text, "@") || strings.Contains(text, "t.me/") {
		return processChannelID(c, db, userID, text)
	}

	// بررسی زمان دلخواه
	if strings.Contains(text, ":") && len(text) == 5 {
		_, err := time.Parse("15:04", text)
		if err == nil {
			return saveScheduleTime(c, db, userID, text)
		}
	}

	// اگر متن طولانی است، احتمالاً پرامپت کانال است
	if len(text) > 20 {
		return processChannelPrompt(c, db, userID, text)
	}

	return c.Send("❌ دستور نامعتبر. لطفاً از منو استفاده کنید.")
}

// پردازش آیدی کانال
func processChannelID(c telebot.Context, db *sql.DB, userID int64, channelInput string) error {
	// استخراج آیدی کانال از متن ورودی
	channelID := extractChannelID(channelInput)
	if channelID == "" {
		return c.Send("❌ آیدی کانال نامعتبر است.")
	}

	// بررسی ادمین بودن ربات در کانال
	isAdmin, err := checkBotAdminStatus(c.Bot(), channelID)
	if err != nil {
		return c.Send("❌ خطا در بررسی کانال. مطمئن شوید کانال وجود دارد و ربات ادمین است.")
	}

	if !isAdmin {
		return c.Send("❌ ربات در کانال ادمین نیست. لطفاً ابتدا ربات را ادمین کنید.")
	}

	// دریافت اطلاعات کانال
	channelTitle, err := getChannelTitle(c.Bot(), channelID)
	if err != nil {
		channelTitle = channelID
	}

	// ذخیره تنظیمات کانال
	err = saveChannelConfig(db, userID, channelID, channelTitle)
	if err != nil {
		return c.Send("❌ خطا در ذخیره تنظیمات کانال")
	}

	return c.Send(fmt.Sprintf(
		"✅ کانال «%s» با موفقیت تنظیم شد.\n\n"+
			"حالا می‌توانید:\n"+
			"• پرامپت مخصوص کانال را تنظیم کنید\n"+
			"• زمان انتشار را مشخص کنید\n"+
			"• کانال را فعال کنید",
		channelTitle,
	))
}

// پردازش پرامپت کانال
func processChannelPrompt(c telebot.Context, db *sql.DB, userID int64, prompt string) error {
	err := updateChannelPrompt(db, userID, prompt)
	if err != nil {
		return c.Send("❌ خطا در ذخیره پرامپت")
	}

	return c.Send("✅ پرامپت کانال با موفقیت ذخیره شد.\n\nاکنون می‌توانید کانال را فعال کنید.")
}

// استخراج آیدی کانال از متن ورودی
func extractChannelID(input string) string {
	input = strings.TrimSpace(input)
	
	// اگر با @ شروع شده
	if strings.HasPrefix(input, "@") {
		return input
	}
	
	// اگر لینک است
	if strings.Contains(input, "t.me/") {
		parts := strings.Split(input, "t.me/")
		if len(parts) > 1 {
			channel := strings.Trim(parts[1], "/")
			if !strings.Contains(channel, "/") {
				return "@" + channel
			}
		}
	}
	
	return ""
}

// بررسی ادمین بودن ربات در کانال
func checkBotAdminStatus(bot *telebot.Bot, channelID string) (bool, error) {
	chat, err := bot.ChatByUsername(channelID)
	if err != nil {
		return false, err
	}

	member, err := bot.ChatMemberOf(chat, &telebot.User{ID: bot.Me.ID})
	if err != nil {
		return false, err
	}

	return member.Role == telebot.Administrator || member.Role == telebot.Creator, nil
}

// دریافت عنوان کانال
func getChannelTitle(bot *telebot.Bot, channelID string) (string, error) {
	chat, err := bot.ChatByUsername(channelID)
	if err != nil {
		return "", err
	}
	return chat.Title, nil
}

// توابع دیتابیس برای مدیریت کانال‌ها
func getChannelConfig(db *sql.DB, userID int64) (*ChannelConfig, error) {
	// این تابع بعداً با جدول کانال‌ها تکمیل می‌شود
	return nil, nil
}

func saveChannelConfig(db *sql.DB, userID int64, channelID, channelTitle string) error {
	// این تابع بعداً با جدول کانال‌ها تکمیل می‌شود
	return nil
}

func updateChannelPrompt(db *sql.DB, userID int64, prompt string) error {
	// این تابع بعداً با جدول کانال‌ها تکمیل می‌شود
	return nil
}

func updateChannelSchedule(db *sql.DB, userID int64, scheduleTime string) error {
	// این تابع بعداً با جدول کانال‌ها تکمیل می‌شود
	return nil
}

func updateChannelPosts(db *sql.DB, userID int64, posts int) error {
	// این تابع بعداً با جدول کانال‌ها تکمیل می‌شود
	return nil
}

func updateChannelStatus(db *sql.DB, userID int64, isActive bool) error {
	// این تابع بعداً با جدول کانال‌ها تکمیل می‌شود
	return nil
}

// تولید و انتشار محتوا در کانال
func GenerateAndPostChannelContent(bot *telebot.Bot, db *sql.DB) {
	// این تابع توسط scheduler فراخوانی می‌شود
	// دریافت تمام کانال‌های فعال
	// تولید محتوا برای هر کانال
	// انتشار در کانال‌ها
}

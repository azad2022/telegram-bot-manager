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

// HandlePrivateMessage مدیریت پیام‌های چت خصوصی
func HandlePrivateMessage(bot *telebot.Bot, db *sql.DB) {
	// منوی اصلی
	bot.Handle("/start", func(c telebot.Context) error {
		user := c.Sender()
		
		// ایجاد کاربر در دیتابیس اگر وجود ندارد
		err := models.CreateUser(db, user.ID, user.Username, user.FirstName, user.LastName)
		if err != nil {
			log.Printf("خطا در ایجاد کاربر: %v", err)
		}

		// بررسی referral
		if len(c.Args()) > 0 {
			handleReferral(c, db)
		}

		return sendMainMenu(c, db)
	})

	// منوی اصلی با دکمه
	bot.Handle("🔙 بازگشت به منوی اصلی", func(c telebot.Context) error {
		return sendMainMenu(c, db)
	})

	// مدیریت پرامپت‌ها
	bot.Handle("🧠 مدیریت پرامپت‌ها", func(c telebot.Context) error {
		return sendPromptManagementMenu(c, db)
	})

	// مدیریت API
	bot.Handle("🔑 مدیریت API", func(c telebot.Context) error {
		return sendAPIManagementMenu(c, db)
	})

	// مشاهده مصرف
	bot.Handle("📊 مشاهده مصرف", func(c telebot.Context) error {
		return showUsageStats(c, db)
	})

	// تنظیمات مدل
	bot.Handle("⚙️ تنظیمات مدل", func(c telebot.Context) error {
		return sendModelSettingsMenu(c, db)
	})

	// تنظیمات کانال
	bot.Handle("🔧 تنظیمات کانال", func(c telebot.Context) error {
		return sendChannelSettingsMenu(c, db)
	})

	// تنظیمات گروه
	bot.Handle("🔨 تنظیمات گروه", func(c telebot.Context) error {
		return sendGroupSettingsMenu(c, db)
	})

	// امتیازگیری
	bot.Handle("🎯 امتیازگیری", func(c telebot.Context) error {
		return sendInvitationMenu(c, db)
	})

	// راهنمای ربات
	bot.Handle("📣 راهنمای ربات", func(c telebot.Context) error {
		return sendHelpGuide(c)
	})

	// پنل مدیریت (فقط برای سازنده)
	bot.Handle("🛠️ پنل مدیریت", func(c telebot.Context) error {
		if c.Sender().ID != 269758292 {
			return c.Send("⛔ دسترسی denied")
		}
		return sendAdminPanel(c, db)
	})
}

// ارسال منوی اصلی
func sendMainMenu(c telebot.Context, db *sql.DB) error {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
	
	user, err := models.GetUserByTelegramID(db, c.Sender().ID)
	if err != nil {
		return c.Send("خطا در دریافت اطلاعات کاربر")
	}

	// ایجاد دکمه‌های منوی اصلی
	btnPrompts := menu.Text("🧠 مدیریت پرامپت‌ها")
	btnAPI := menu.Text("🔑 مدیریت API")
	btnUsage := menu.Text("📊 مشاهده مصرف")
	btnModel := menu.Text("⚙️ تنظیمات مدل")
	btnChannel := menu.Text("🔧 تنظیمات کانال")
	btnGroup := menu.Text("🔨 تنظیمات گروه")
	btnInvite := menu.Text("🎯 امتیازگیری")
	btnHelp := menu.Text("📣 راهنمای ربات")

	// چیدمان منو
	rows := []telebot.Row{
		menu.Row(btnPrompts, btnAPI),
		menu.Row(btnUsage, btnModel),
		menu.Row(btnChannel, btnGroup),
		menu.Row(btnInvite, btnHelp),
	}

	// اگر سازنده هست، پنل مدیریت اضافه کن
	if c.Sender().ID == 269758292 {
		btnAdmin := menu.Text("🛠️ پنل مدیریت")
		rows = append(rows, menu.Row(btnAdmin))
	}

	menu.Reply(rows...)

	welcomeMsg := "🤖 به ربات ChatGPT خوش آمدید!\n\n"
	
	if user != nil && user.IsVIP {
		welcomeMsg += "👑 وضعیت: VIP کاربر\n"
		if user.VIPUntil.Valid {
			welcomeMsg += fmt.Sprintf("📅 اعتبار VIP: %s\n", user.VIPUntil.Time.Format("2006-01-02"))
		}
	} else {
		welcomeMsg += "🔹 وضعیت: کاربر عادی\n"
	}

	welcomeMsg += "\nاز منوی زیر انتخاب کنید:"

	return c.Send(welcomeMsg, menu)
}

// مدیریت پرامپت‌ها
func sendPromptManagementMenu(c telebot.Context, db *telebot.Context) error {
	userID := c.Sender().ID
	user, err := models.GetUserByTelegramID(db, userID)
	if err != nil {
		return c.Send("خطا در دریافت اطلاعات کاربر")
	}

	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
	
	// دریافت پرامپت‌های کاربر
	prompts, err := models.GetUserPrompts(db, userID)
	if err != nil {
		return c.Send("خطا در دریافت پرامپت‌ها")
	}

	activePrompt, _ := models.GetActivePrompt(db, userID)

	// ایجاد دکمه‌های مدیریت پرامپت
	btnAdd := menu.Text("➕ افزودن پرامپت جدید")
	btnList := menu.Text("📋 لیست پرامپت‌های من")
	btnBack := menu.Text("🔙 بازگشت به منوی اصلی")

	menu.Reply(
		menu.Row(btnAdd),
		menu.Row(btnList),
		menu.Row(btnBack),
	)

	message := "🧠 مدیریت پرامپت‌ها\n\n"
	
	if activePrompt != nil {
		message += fmt.Sprintf("🟢 پرامپت فعال: %s\n", activePrompt.Title)
	} else {
		message += "🔴 هیچ پرامپت فعالی ندارید\n"
	}

	count := len(prompts)
	var maxPrompts int
	if user != nil && user.IsVIP {
		maxPrompts = 10
	} else {
		maxPrompts = 3
	}

	message += fmt.Sprintf("📝 تعداد: %d/%d پرامپت\n\n", count, maxPrompts)
	message += "از گزینه‌های زیر انتخاب کنید:"

	// هندلر برای افزودن پرامپت جدید
	bot.Handle("➕ افزودن پرامپت جدید", func(c telebot.Context) error {
		return handleAddNewPrompt(c, db, user)
	})

	// هندلر برای لیست پرامپت‌ها
	bot.Handle("📋 لیست پرامپت‌های من", func(c telebot.Context) error {
		return handlePromptList(c, db, prompts)
	})

	return c.Send(message, menu)
}

// مدیریت API
func sendAPIManagementMenu(c telebot.Context, db *sql.DB) error {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnAdd := menu.Text("🔑 افزودن کلید API جدید")
	btnList := menu.Text("📋 لیست کلیدهای من")
	btnUsage := menu.Text("📊 مشاهده مصرف")
	btnBack := menu.Text("🔙 بازگشت به منوی اصلی")

	menu.Reply(
		menu.Row(btnAdd, btnList),
		menu.Row(btnUsage),
		menu.Row(btnBack),
	)

	// هندلر برای افزودن API Key
	bot.Handle("🔑 افزودن کلید API جدید", func(c telebot.Context) error {
		return c.Send("لطفاً کلید API جدید خود را وارد کنید:\n\nفرمت: sk-...")
	})

	// هندلر برای لیست API Keys
	bot.Handle("📋 لیست کلیدهای من", func(c telebot.Context) error {
		return handleAPIKeyList(c, db)
	})

	// هندلر برای مشاهده مصرف
	bot.Handle("📊 مشاهده مصرف", func(c telebot.Context) error {
		return handleUsageDetails(c, db)
	})

	return c.Send("🔑 مدیریت API Keys\n\nاز گزینه‌های زیر انتخاب کنید:", menu)
}

// مشاهده آمار مصرف
func showUsageStats(c telebot.Context, db *sql.DB) error {
	userID := c.Sender().ID

	daily, weekly, monthly, err := models.GetUsageStats(db, userID)
	if err != nil {
		return c.Send("خطا در دریافت آمار مصرف")
	}

	user, err := models.GetUserByTelegramID(db, userID)
	if err != nil {
		return c.Send("خطا در دریافت اطلاعات کاربر")
	}

	withinLimit, remaining, _ := models.CheckUsageLimit(db, userID, user != nil && user.IsVIP)

	message := "📊 آمار مصرف شما\n\n"
	message += fmt.Sprintf("📅 مصرف امروز: %d توکن\n", daily)
	message += fmt.Sprintf("📈 مصرف این هفته: %d توکن\n", weekly)
	message += fmt.Sprintf("📋 مصرف این ماه: %d توکن\n\n", monthly)

	if withinLimit {
		message += fmt.Sprintf("✅ وضعیت: عادی (%d توکن باقی‌مانده)\n", remaining)
	} else {
		message += "⚠️  شما به سقف مصرف روزانه رسیده‌اید\n"
	}

	if user != nil && !user.IsVIP {
		message += "\n🎯 برای مصرف نامحدود به VIP ارتقا پیدا کنید!"
	}

	return c.Send(message)
}

// هندلر referral
func handleReferral(c telebot.Context, db *sql.DB) {
	args := c.Args()
	if len(args) == 0 {
		return
	}

	refParam := args[0]
	if strings.HasPrefix(refParam, "ref_") {
		referrerIDStr := strings.TrimPrefix(refParam, "ref_")
		referrerID, err := strconv.ParseInt(referrerIDStr, 10, 64)
		if err == nil {
			// افزایش تعداد دعوت‌های referrer
			models.IncrementInviteCount(db, referrerID)
			
			// بررسی و فعال‌سازی VIP در صورت نیاز
			user, err := models.GetUserByTelegramID(db, referrerID)
			if err == nil && user != nil {
				if user.InviteCount >= 20 {
					models.ActivateVIP(db, referrerID, 1) // 1 روز VIP
					
					// ارسال پیام تبریک به دعوت‌کننده
					bot.Send(&telebot.User{ID: referrerID}, 
						"🎉 تبریک! شما به ۲۰ دعوت موفق رسیدید!\n" +
						"🎁 ۱ روز VIP برای شما فعال شد!")
				}
			}
		}
	}
}

// ارسال راهنمای ربات
func sendHelpGuide(c telebot.Context) error {
	guide := `📣 راهنمای ربات ChatGPT

🔸 نحوه استفاده در گروه:
پیام خود را با * شروع کنید
مثال: *سلام چطوری می‌تونم انگلیسی یاد بگیرم؟

🔸 در چت خصوصی:
از منو برای مدیریت پرامپت‌ها و API استفاده کنید

🔸 امکانات کاربران عادی:
• ۳ پرامپت شخصی
• استفاده از API شخصی
• ۱۰,۰۰۰ توکن در روز

🔸 امکانات کاربران VIP:
• ۱۰ پرامپت شخصی  
• دسترسی به GPT-4
• مصرف نامحدود
• تولید محتوای خودکار در کانال

🎯 با دعوت دوستان، امتیاز بگیرید و رایگان VIP شوید!`

	return c.Send(guide)
}

// Note: توابع کمکی دیگر مانند handleAddNewPrompt, handlePromptList, 
// handleAPIKeyList و... در ادامه پیاده‌سازی خواهند شد

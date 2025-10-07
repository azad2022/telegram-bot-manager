package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/telebot.v3"

	"telegram-bot-manager/database"
	"telegram-bot-manager/models"
	"telegram-bot-manager/services"
)

// HandleGroupMessages مدیریت پیام‌های گروه
func HandleGroupMessages(bot *telebot.Bot, db *sql.DB) {
	// پردازش پیام‌هایی که با * شروع می‌شوند
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		// فقط در گروه‌ها پردازش کن
		if c.Chat().Type != telebot.ChatGroup && c.Chat().Type != telebot.ChatSuperGroup {
			return nil
		}

		text := c.Text()
		if !strings.HasPrefix(text, "*") {
			return nil
		}

		return handleGroupQuestion(bot, c, db, text)
	})
}

// پردازش سوالات گروه
func handleGroupQuestion(bot *telebot.Bot, c telebot.Context, db *sql.DB, question string) error {
	user := c.Sender()
	chat := c.Chat()
	
	// حذف * از ابتدای سوال
	question = strings.TrimPrefix(question, "*")
	question = strings.TrimSpace(question)

	if question == "" {
		return c.Reply("لطفاً پس از * سوال خود را وارد کنید.")
	}

	// بررسی rate limiting
	canProceed, currentCount, err := checkGroupRateLimit(chat.ID)
	if err != nil {
		log.Printf("خطا در بررسی rate limit: %v", err)
		return c.Reply("خطای سیستمی. لطفاً مجدد تلاش کنید.")
	}

	if !canProceed {
		// فقط یک بار هشدار بده
		warningSent, err := database.IsWarningSent(fmt.Sprintf("%d", chat.ID))
		if err == nil && !warningSent {
			database.SetWarningSent(fmt.Sprintf("%d", chat.ID))
			
			menu := &telebot.ReplyMarkup{}
			btnVIP := menu.URL("🎯 ارتقاء به VIP", "https://t.me/gpt_yourbot?start=vip_request")
			menu.Inline(menu.Row(btnVIP))
			
			message := fmt.Sprintf(
				"⚠️ محدودیت پاسخ‌گویی در گروه فعال است.\n"+
					"حداکثر ۵ سوال در دقیقه پاسخ داده می‌شود.\n"+
					"لطفاً چند لحظه بعد دوباره تلاش کنید.\n\n"+
					"🕒 سوالات این دقیقه: %d/۵",
				currentCount,
			)
			
			return c.Reply(message, menu)
		}
		return nil
	}

	// افزایش شمارنده rate limit
	database.IncrementGroupRateLimit(fmt.Sprintf("%d", chat.ID), time.Minute)

	// نشان دادن تایپینگ
	bot.Notify(chat, telebot.Typing)

	// دریافت اطلاعات کاربر
	dbUser, err := models.GetUserByTelegramID(db, user.ID)
	if err != nil {
		log.Printf("خطا در دریافت کاربر: %v", err)
		return c.Reply("خطا در دریافت اطلاعات کاربر.")
	}

	// دریافت پرامپت فعال کاربر
	var promptContent string
	activePrompt, err := models.GetActivePrompt(db, user.ID)
	if err == nil && activePrompt != nil {
		promptContent = activePrompt.Content
	} else {
		// پرامپت پیش‌فرض
		promptContent = "تو یک دستیار هوشمند هستی. به سوالات کاربران به صورت مفید و دقیق پاسخ بده."
	}

	// دریافت API Key کاربر
	apiKey, err := models.GetActiveAPIKey(db, user.ID)
	if err != nil || apiKey == nil {
		menu := &telebot.ReplyMarkup{}
		btnAPI := menu.URL("🔑 تنظیم API", "https://t.me/gpt_yourbot?start=api_setup")
		menu.Inline(menu.Row(btnAPI))
		
		return c.Reply(
			"🔑 شما هنوز API Key خود را تنظیم نکرده‌اید.\n"+
				"لطفاً در چت خصوصی با ربات، API Key خود را اضافه کنید.",
			menu,
		)
	}

	// بررسی سقف مصرف
	if dbUser != nil {
		withinLimit, remaining, err := models.CheckUsageLimit(db, user.ID, dbUser.IsVIP)
		if err == nil && !withinLimit && !dbUser.IsVIP {
			menu := &telebot.ReplyMarkup{}
			btnVIP := menu.URL("🎯 ارتقاء به VIP", "https://t.me/gpt_yourbot?start=vip_request")
			menu.Inline(menu.Row(btnVIP))
			
			return c.Reply(
				"⚠️ شما به سقف مصرف روزانه رسیده‌اید.\n"+
					"برای استفاده نامحدود به VIP ارتقا پیدا کنید.",
				menu,
			)
		}
	}

	// ارسال به ChatGPT
	response, tokensUsed, err := services.CallChatGPT(apiKey.APIKey, promptContent, question, dbUser != nil && dbUser.IsVIP)
	if err != nil {
		log.Printf("خطا در تماس با ChatGPT: %v", err)
		
		if strings.Contains(err.Error(), "insufficient_quota") {
			return c.Reply("❌ سقف مصرف API Key شما به پایان رسیده است. لطفاً API Key جدیدی اضافه کنید.")
		} else if strings.Contains(err.Error(), "invalid_api_key") {
			return c.Reply("❌ API Key نامعتبر است. لطفاً API Key خود را بررسی کنید.")
		}
		
		return c.Reply("❌ خطا در ارتباط با سرویس ChatGPT. لطفاً مجدد تلاش کنید.")
	}

	// ثبت مصرف توکن
	if tokensUsed > 0 {
		cost := float64(tokensUsed) * 0.002 / 1000 // تقریباً 0.002 دلار per 1K tokens
		models.RecordTokenUsage(db, user.ID, tokensUsed, cost)
	}

	// اضافه کردن متن پایانی اگر کاربر VIP است و تنظیم کرده
	finalResponse := response
	if dbUser != nil && dbUser.IsVIP {
		// در اینجا می‌توان متن پایانی از تنظیمات گروه را اضافه کرد
		// finalResponse = response + "\n\n" + groupSettings.FooterText
	}

	// اضافه کردن دکمه ارتقا برای کاربران عادی
	var replyMarkup *telebot.ReplyMarkup
	if dbUser == nil || !dbUser.IsVIP {
		replyMarkup = &telebot.ReplyMarkup{}
		btnVIP := replyMarkup.URL("🎯 ارتقاء به VIP", "https://t.me/gpt_yourbot?start=vip_request")
		replyMarkup.Inline(replyMarkup.Row(btnVIP))
	}

	return c.Reply(finalResponse, replyMarkup)
}

// بررسی rate limit گروه
func checkGroupRateLimit(chatID int64) (bool, int, error) {
	key := fmt.Sprintf("%d", chatID)
	
	// دریافت تعداد فعلی
	currentCount, err := database.GetGroupRateLimit(key)
	if err != nil {
		return false, 0, err
	}

	// کاربران VIP محدودیت ندارند (در اینجا می‌توان بررسی کرد که کاربر VIP است یا نه)
	// برای سادگی، در این نسخه همه کاربران در گروه محدودیت یکسان دارند
	
	if currentCount >= 5 { // حداکثر ۵ سوال در دقیقه
		return false, currentCount, nil
	}

	return true, currentCount, nil
}

// پردازش چند سوال همزمان
func handleMultipleQuestions(bot *telebot.Bot, c telebot.Context, db *sql.DB, questions map[int64]string) error {
	chat := c.Chat()
	
	// بررسی rate limiting برای سوالات چندگانه
	canProceed, currentCount, err := checkGroupRateLimit(chat.ID)
	if err != nil || !canProceed {
		return nil // سکوت در صورت محدودیت
	}

	// افزایش شمارنده برای هر سوال
	for range questions {
		database.IncrementGroupRateLimit(fmt.Sprintf("%d", chat.ID), time.Minute)
	}

	// نشان دادن تایپینگ
	bot.Notify(chat, telebot.Typing)

	// جمع‌آوری پاسخ‌ها
	var responses []string
	totalTokens := 0

	for userID, question := range questions {
		// دریافت اطلاعات کاربر
		dbUser, err := models.GetUserByTelegramID(db, userID)
		if err != nil {
			continue
		}

		// دریافت API Key کاربر
		apiKey, err := models.GetActiveAPIKey(db, userID)
		if err != nil || apiKey == nil {
			responses = append(responses, fmt.Sprintf("👤 کاربر %d: 🔑 API Key تنظیم نشده", userID))
			continue
		}

		// دریافت پرامپت فعال
		var promptContent string
		activePrompt, err := models.GetActivePrompt(db, userID)
		if err == nil && activePrompt != nil {
			promptContent = activePrompt.Content
		} else {
			promptContent = "تو یک دستیار هوشمند هستی. به سوالات کاربران به صورت مفید و دقیق پاسخ بده."
		}

		// ارسال به ChatGPT
		response, tokensUsed, err := services.CallChatGPT(apiKey.APIKey, promptContent, question, dbUser != nil && dbUser.IsVIP)
		if err != nil {
			responses = append(responses, fmt.Sprintf("👤 کاربر %d: ❌ خطا در دریافت پاسخ", userID))
			continue
		}

		// ثبت مصرف توکن
		if tokensUsed > 0 {
			cost := float64(tokensUsed) * 0.002 / 1000
			models.RecordTokenUsage(db, userID, tokensUsed, cost)
			totalTokens += tokensUsed
		}

		// کوتاه کردن پاسخ اگر طولانی باشد
		if len(response) > 500 {
			response = response[:500] + "..."
		}

		responses = append(responses, fmt.Sprintf("👤 کاربر %d: %s", userID, response))
	}

	// ترکیب تمام پاسخ‌ها
	finalResponse := strings.Join(responses, "\n\n")
	
	// اضافه کردن اطلاعات rate limit
	finalResponse += fmt.Sprintf("\n\n🕒 %d/۵ سوال در این دقیقه", len(questions)+currentCount)

	// اضافه کردن دکمه ارتقا اگر کاربران عادی هستند
	replyMarkup := &telebot.ReplyMarkup{}
	btnVIP := replyMarkup.URL("🎯 ارتقاء به VIP", "https://t.me/gpt_yourbot?start=vip_request")
	replyMarkup.Inline(replyMarkup.Row(btnVIP))

	return c.Reply(finalResponse, replyMarkup)
}

// مدیریت اضافه شدن ربات به گروه جدید
func HandleBotAddedToGroup(bot *telebot.Bot, c telebot.Context, db *sql.DB) {
	chat := c.Chat()
	addedBy := c.Sender()

	log.Printf("ربات به گروه اضافه شد: %s (%d) توسط کاربر: %d", chat.Title, chat.ID, addedBy.ID)

	// ارسال پیام خوش‌آمد به صورت خصوصی به کاربر
	welcomeMsg := fmt.Sprintf(
		"🤖 ربات ChatGPT به گروه «%s» اضافه شد!\n\n"+
			"📝 نحوه استفاده در گروه:\n"+
			"• پیام خود را با * شروع کنید\n"+
			"• مثال: *سلام چطوری می‌تونم انگلیسی یاد بگیرم؟\n\n"+
			"⚙️ برای مدیریت تنظیمات گروه:\n"+
			"• به چت خصوصی با ربات مراجعه کنید\n"+
			"• منوی «🔨 تنظیمات گروه» را انتخاب کنید\n\n"+
			"🎯 برای حذف محدودیت‌ها به VIP ارتقا پیدا کنید!",
		chat.Title,
	)

	bot.Send(addedBy, welcomeMsg)

	// ارسال پیام در گروه (اختیاری)
	groupMsg := "🤖 ربات ChatGPT فعال شد!\n\n" +
		"برای استفاده، پیام خود را با * شروع کنید.\n" +
		"مثال: *سوال خود را اینجا بنویسید"

	c.Send(groupMsg)
}

// بررسی وضعیت ربات در گروه
func CheckBotStatusInGroup(bot *telebot.Bot, chatID int64) (bool, error) {
	chat, err := bot.ChatByID(chatID)
	if err != nil {
		return false, err
	}

	// بررسی اینکه ربات در گروه هست و ادمین است
	member, err := bot.ChatMemberOf(chat, &telebot.User{ID: bot.Me.ID})
	if err != nil {
		return false, err
	}

	return member.Role == telebot.Administrator || member.Role == telebot.Creator, nil
}

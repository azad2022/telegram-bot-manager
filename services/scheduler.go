package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/telebot.v3"

	"telegram-bot-manager/database"
	"telegram-bot-manager/models"
)

// Scheduler - سیستم زمان‌بندی تولید محتوا
type Scheduler struct {
	bot *telebot.Bot
	db  *sql.DB
}

// NewScheduler - ایجاد نمونه جدید scheduler
func NewScheduler(bot *telebot.Bot, db *sql.DB) *Scheduler {
	return &Scheduler{
		bot: bot,
		db:  db,
	}
}

// Start - شروع زمان‌بندی
func (s *Scheduler) Start() {
	log.Println("🕒 سیستم زمان‌بندی شروع به کار کرد...")

	// اجرای بررسی فوری
	go s.checkAndPostContent()

	// زمان‌بندی بررسی هر دقیقه
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		go s.checkAndPostContent()
	}
}

// checkAndPostContent - بررسی و انتشار محتوا
func (s *Scheduler) checkAndPostContent() {
	now := time.Now()
	currentTime := now.Format("15:04")

	log.Printf("🔍 بررسی زمان‌بندی برای ساعت: %s", currentTime)

	// دریافت تمام کانال‌های فعال
	activeChannels, err := s.getActiveChannels()
	if err != nil {
		log.Printf("❌ خطا در دریافت کانال‌های فعال: %v", err)
		return
	}

	for _, channel := range activeChannels {
		if channel.ScheduleTime == currentTime {
			go s.processChannelContent(channel)
		}
	}
}

// ChannelConfig - ساختار کانال (مشابه handlers/channel.go)
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

// getActiveChannels - دریافت کانال‌های فعال از دیتابیس
func (s *Scheduler) getActiveChannels() ([]ChannelConfig, error) {
	// TODO: جایگزین با کوئری واقعی هنگامی که جدول channels ایجاد شد
	// در حال حاضر از نمونه‌های تستی استفاده می‌کنیم
	var channels []ChannelConfig

	// این بخش موقتی است - بعداً با جدول واقعی جایگزین می‌شود
	rows, err := s.db.Query(`
		SELECT id, owner_id, channel_id, channel_title, prompt, 
		       schedule_time, posts_per_batch, is_active, created_at
		FROM channels 
		WHERE is_active = true
	`)
	if err != nil {
		// اگر جدول وجود ندارد، نمونه‌های تستی برگردان
		return s.getMockChannels(), nil
	}
	defer rows.Close()

	for rows.Next() {
		var channel ChannelConfig
		err := rows.Scan(
			&channel.ID, &channel.OwnerID, &channel.ChannelID,
			&channel.ChannelTitle, &channel.Prompt, &channel.ScheduleTime,
			&channel.PostsPerBatch, &channel.IsActive, &channel.CreatedAt,
		)
		if err != nil {
			continue
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// getMockChannels - کانال‌های نمونه (موقتی)
func (s *Scheduler) getMockChannels() []ChannelConfig {
	// این تابع موقتی است و بعداً حذف می‌شود
	return []ChannelConfig{
		{
			ID:           1,
			OwnerID:      269758292,
			ChannelID:    "@test_channel",
			ChannelTitle: "کانال تست",
			Prompt:       "تولید محتوای آموزشی در مورد برنامه‌نویسی و تکنولوژی",
			ScheduleTime: "09:00",
			PostsPerBatch: 1,
			IsActive:     true,
			CreatedAt:    time.Now(),
		},
	}
}

// processChannelContent - پردازش محتوای یک کانال
func (s *Scheduler) processChannelContent(channel ChannelConfig) {
	log.Printf("🎯 شروع تولید محتوا برای کانال: %s", channel.ChannelTitle)

	// دریافت API Key مالک کانال
	apiKey, err := models.GetActiveAPIKey(s.db, channel.OwnerID)
	if err != nil || apiKey == nil {
		s.notifyOwner(channel.OwnerID, 
			"❌ خطا در تولید محتوای خودکار\n" +
			"دلیل: API Key تنظیم نشده است\n" +
			"کانال: " + channel.ChannelTitle)
		return
	}

	// بررسی ادمین بودن ربات در کانال
	isAdmin, err := s.checkBotAdminStatus(channel.ChannelID)
	if err != nil || !isAdmin {
		s.notifyOwner(channel.OwnerID,
			"❌ خطا در تولید محتوای خودکار\n" +
			"دلیل: ربات در کانال ادمین نیست\n" +
			"کانال: " + channel.ChannelTitle)
		return
	}

	// تولید محتوا
	for i := 0; i < channel.PostsPerBatch; i++ {
		content, tokensUsed, err := s.generateChannelContent(apiKey.APIKey, channel.Prompt)
		if err != nil {
			log.Printf("❌ خطا در تولید محتوا برای کانال %s: %v", channel.ChannelTitle, err)
			s.notifyOwner(channel.OwnerID,
				"❌ خطا در تولید محتوای خودکار\n" +
				"دلیل: " + err.Error() + "\n" +
				"کانال: " + channel.ChannelTitle)
			continue
		}

		// انتشار محتوا در کانال
		err = s.postToChannel(channel.ChannelID, content)
		if err != nil {
			log.Printf("❌ خطا در انتشار محتوا در کانال %s: %v", channel.ChannelTitle, err)
			s.notifyOwner(channel.OwnerID,
				"❌ خطا در انتشار محتوای خودکار\n" +
				"دلیل: " + err.Error() + "\n" +
				"کانال: " + channel.ChannelTitle)
			continue
		}

		// ثبت مصرف توکن
		if tokensUsed > 0 {
			cost := CalculateCost(tokensUsed, true) // کاربران VIP هستند
			models.RecordTokenUsage(s.db, channel.OwnerID, tokensUsed, cost)
		}

		log.Printf("✅ محتوا با موفقیت در کانال %s منتشر شد (%d توکن)", 
			channel.ChannelTitle, tokensUsed)

		// تأثیر بین پست‌ها
		if i < channel.PostsPerBatch-1 {
			time.Sleep(2 * time.Second)
		}
	}

	// اطلاع‌رسانی موفقیت
	s.notifyOwner(channel.OwnerID,
		"✅ محتوای خودکار با موفقیت منتشر شد\n" +
		"کانال: " + channel.ChannelTitle + "\n" +
		"تعداد پست: " + fmt.Sprintf("%d", channel.PostsPerBatch))
}

// generateChannelContent - تولید محتوا برای کانال
func (s *Scheduler) generateChannelContent(apiKey, prompt string) (string, int, error) {
	systemPrompt := fmt.Sprintf(
		"تو یک تولیدکننده محتوای حرفه‌ای برای کانال‌های تلگرام هستی.\n" +
		"محتوایی تولید کن که:\n" +
		"- جذاب و مفید باشد\n" +
		"- حدود ۲۰۰-۳۰۰ کلمه باشد\n" +
		"- برای انتشار در کانال مناسب باشد\n" +
		"- دارای ساختار منظم\n" +
		"- حاوی نکات کاربردی\n\n" +
		"دستورالعمل خاص: %s",
		prompt,
	)

	userMessage := "لطفاً یک پست جذاب برای کانل تلگرام تولید کن."

	return GenerateChannelContent(apiKey, userMessage, 250) // حدود ۲۵۰ کلمه
}

// postToChannel - انتشار محتوا در کانال
func (s *Scheduler) postToChannel(channelID, content string) error {
	chat, err := s.bot.ChatByUsername(channelID)
	if err != nil {
		return fmt.Errorf("یافتن کانال: %v", err)
	}

	// اضافه کردن هشتگ و فرمت‌بندی
	formattedContent := formatChannelPost(content)

	_, err = s.bot.Send(chat, formattedContent, &telebot.SendOptions{
		ParseMode: telebot.ModeHTML,
	})
	if err != nil {
		return fmt.Errorf("ارسال پیام: %v", err)
	}

	return nil
}

// formatChannelPost - فرمت‌بندی پست کانال
func formatChannelPost(content string) string {
	// اضافه کردن هشتگ‌های مرتبط
	hashtags := "\n\n#آموزش #تکنولوژی #برنامه‌نویسی #هوش_مصنوعی"
	
	// فرمت‌بندی HTML ساده
	formatted := fmt.Sprintf(
		"📚 <b>نکته آموزشی</b>\n\n%s%s\n\n🤖 <i>تولید شده توسط ChatGPT</i>",
		content, hashtags,
	)

	return formatted
}

// checkBotAdminStatus - بررسی ادمین بودن ربات در کانال
func (s *Scheduler) checkBotAdminStatus(channelID string) (bool, error) {
	chat, err := s.bot.ChatByUsername(channelID)
	if err != nil {
		return false, err
	}

	member, err := s.bot.ChatMemberOf(chat, &telebot.User{ID: s.bot.Me.ID})
	if err != nil {
		return false, err
	}

	return member.Role == telebot.Administrator || member.Role == telebot.Creator, nil
}

// notifyOwner - اطلاع‌رسانی به مالک کانال
func (s *Scheduler) notifyOwner(ownerID int64, message string) {
	user := &telebot.User{ID: ownerID}
	_, err := s.bot.Send(user, message)
	if err != nil {
		log.Printf("❌ خطا در اطلاع‌رسانی به کاربر %d: %v", ownerID, err)
	}
}

// CheckVIPExpirations - بررسی انقضای اشتراک‌های VIP
func (s *Scheduler) CheckVIPExpirations() {
	log.Println("🔍 بررسی انقضای اشتراک‌های VIP...")

	err := models.CheckVIPExpiration(s.db)
	if err != nil {
		log.Printf("❌ خطا در بررسی انقضای VIP: %v", err)
		return
	}

	// اطلاع‌رسانی به کاربرانی که VIP آنها در حال اتمام است
	s.notifyExpiringVIPs()
}

// notifyExpiringVIPs - اطلاع‌رسانی به کاربران در حال اتمام VIP
func (s *Scheduler) notifyExpiringVIPs() {
	// کاربرانی که VIP آنها تا ۳ روز دیگر منقضی می‌شود
	rows, err := s.db.Query(`
		SELECT telegram_id, username, first_name, vip_until 
		FROM users 
		WHERE is_vip = true 
		AND vip_until BETWEEN NOW() AND NOW() + INTERVAL '3 days'
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		var username, firstName string
		var vipUntil time.Time

		err := rows.Scan(&userID, &username, &firstName, &vipUntil)
		if err != nil {
			continue
		}

		daysLeft := int(vipUntil.Sub(time.Now()).Hours() / 24)
		message := fmt.Sprintf(
			"⏳ اعتبار VIP شما در حال اتمام است\n\n" +
			"📅 تا انقضا: %d روز\n" +
			"✅ برای تمدید، از منوی «🎯 امتیازگیری» استفاده کنید\n\n" +
			"با تمدید از امکانات ویژه بهره‌مند شوید!",
			daysLeft,
		)

		s.notifyOwner(userID, message)
	}
}

// CleanupOldData - پاک‌سازی داده‌های قدیمی
func (s *Scheduler) CleanupOldData() {
	log.Println("🧹 پاک‌سازی داده‌های قدیمی...")

	// پاک‌سازی لاگ‌های قدیمی Redis (بیش از ۷ روز)
	s.cleanupRedisData()

	// پاک‌سازی داده‌های موقت
	s.cleanupTempData()
}

// cleanupRedisData - پاک‌سازی داده‌های قدیمی Redis
func (s *Scheduler) cleanupRedisData() {
	// این تابع بعداً با Redis Keys مشخص تکمیل می‌شود
	log.Println("✅ پاک‌سازی داده‌های موقت انجام شد")
}

// cleanupTempData - پاک‌سازی داده‌های موقت
func (s *Scheduler) cleanupTempData() {
	// پاک‌سازی لاگ‌های قدیمی و داده‌های موقت
	// این تابع می‌تواند گسترش یابد
}

// StartMaintenance - شروع عملیات نگهداری
func (s *Scheduler) StartMaintenance() {
	log.Println("🔧 شروع عملیات نگهداری سیستم...")

	// اجرای وظایف نگهداری
	go s.periodicMaintenance()
}

// periodicMaintenance - عملیات دوره‌ای نگهداری
func (s *Scheduler) periodicMaintenance() {
	// بررسی انقضای VIP هر روز
	vipTicker := time.NewTicker(24 * time.Hour)
	defer vipTicker.Stop()

	// پاک‌سازی داده‌ها هر ۷ روز
	cleanupTicker := time.NewTicker(7 * 24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-vipTicker.C:
			s.CheckVIPExpirations()
		case <-cleanupTicker.C:
			s.CleanupOldData()
		}
	}
}

// GetSchedulerStatus - دریافت وضعیت زمان‌بندی
func (s *Scheduler) GetSchedulerStatus() string {
	activeChannels, err := s.getActiveChannels()
	if err != nil {
		return "❌ خطا در دریافت وضعیت"
	}

	status := fmt.Sprintf(
		"🕒 وضعیت زمان‌بندی\n\n" +
		"🔸 کانال‌های فعال: %d\n" +
		"🔸 آخرین بررسی: %s\n" +
		"🔸 وضعیت: 🟢 در حال اجرا\n\n",
		len(activeChannels), time.Now().Format("15:04:05"),
	)

	if len(activeChannels) > 0 {
		status += "📋 کانال‌های فعال:\n"
		for _, channel := range activeChannels {
			status += fmt.Sprintf("• %s - %s\n", channel.ChannelTitle, channel.ScheduleTime)
		}
	}

	return status
}

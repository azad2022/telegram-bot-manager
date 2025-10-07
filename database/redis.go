package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client
var ctx = context.Background()

func InitRedis(redisURL, password string) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: password,
		DB:       0, // use default DB
	})

	// تست اتصال
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("خطا در اتصال به Redis: %v", err)
	}

	log.Println("✅ اتصال به Redis برقرار شد")
	return nil
}

// توابع کمکی برای مدیریت داده‌های موقت

// ذخیره rate limit برای گروه
func SetGroupRateLimit(groupID string, count int, expiration time.Duration) error {
	key := fmt.Sprintf("rate_limit:%s", groupID)
	return RDB.Set(ctx, key, count, expiration).Err()
}

// دریافت rate limit برای گروه
func GetGroupRateLimit(groupID string) (int, error) {
	key := fmt.Sprintf("rate_limit:%s", groupID)
	val, err := RDB.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil // اگر وجود نداشته باشد
	}
	return val, err
}

// افزایش rate limit برای گروه
func IncrementGroupRateLimit(groupID string, expiration time.Duration) (int, error) {
	key := fmt.Sprintf("rate_limit:%s", groupID)
	val, err := RDB.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	
	// تنظیم expiration اگر اولین بار است
	if val == 1 {
		RDB.Expire(ctx, key, expiration)
	}
	
	return int(val), nil
}

// ذخیره پرامپت فعال کاربر
func SetUserActivePrompt(userID int64, promptID int) error {
	key := fmt.Sprintf("user_prompt:%d", userID)
	return RDB.Set(ctx, key, promptID, 24*time.Hour).Err()
}

// دریافت پرامپت فعال کاربر
func GetUserActivePrompt(userID int64) (int, error) {
	key := fmt.Sprintf("user_prompt:%d", userID)
	val, err := RDB.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil // اگر وجود نداشته باشد
	}
	return val, err
}

// مدیریت هشدارهای ارسال شده
func SetWarningSent(groupID string) error {
	key := fmt.Sprintf("warning_sent:%s", groupID)
	return RDB.Set(ctx, key, true, time.Minute).Err()
}

// بررسی اینکه آیا هشدار ارسال شده
func IsWarningSent(groupID string) (bool, error) {
	key := fmt.Sprintf("warning_sent:%s", groupID)
	val, err := RDB.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// مدیریت زمان‌بندی کانال‌ها
func SetChannelNextPost(channelID string, nextTime time.Time) error {
	key := fmt.Sprintf("channel_next_post:%s", channelID)
	return RDB.Set(ctx, key, nextTime.Unix(), 0).Err()
}

func GetChannelNextPost(channelID string) (time.Time, error) {
	key := fmt.Sprintf("channel_next_post:%s", channelID)
	val, err := RDB.Get(ctx, key).Int64()
	if err == redis.Nil {
		return time.Time{}, nil // اگر وجود نداشته باشد
	}
	return time.Unix(val, 0), nil
}

// مدیریت دعوت‌ها
func SetUserInvitationCount(userID int64, count int) error {
	key := fmt.Sprintf("invite_count:%d", userID)
	return RDB.Set(ctx, key, count, 30*24*time.Hour).Err() // 30 روز
}

func GetUserInvitationCount(userID int64) (int, error) {
	key := fmt.Sprintf("invite_count:%d", userID)
	val, err := RDB.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil // اگر وجود نداشته باشد
	}
	return val, err
}

func IncrementUserInvitationCount(userID int64) (int, error) {
	key := fmt.Sprintf("invite_count:%d", userID)
	val, err := RDB.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	
	// تنظیم expiration اگر اولین بار است
	if val == 1 {
		RDB.Expire(ctx, key, 30*24*time.Hour)
	}
	
	return int(val), nil
}

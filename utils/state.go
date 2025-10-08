package utils

import (
	"context"
	"fmt"
	"time"

	"telegram-bot-manager/database"
)

// RedisStateManager مدیریت state کاربران در Redis
type RedisStateManager struct{}

// SetState تنظیم وضعیت (state) موقت کاربر
func (r *RedisStateManager) SetState(ctx context.Context, userID int64, state string, ttl time.Duration) error {
	key := fmt.Sprintf("state:%d", userID)
	return database.RDB.Set(ctx, key, state, ttl).Err()
}

// GetState دریافت وضعیت فعلی کاربر
func (r *RedisStateManager) GetState(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("state:%d", userID)
	return database.RDB.Get(ctx, key).Result()
}

// ClearState حذف وضعیت کاربر
func (r *RedisStateManager) ClearState(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("state:%d", userID)
	return database.RDB.Del(ctx, key).Err()
}

// Helper instance عمومی برای استفاده در سایر پکیج‌ها
var State = &RedisStateManager{}

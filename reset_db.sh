#!/bin/bash
set -e

echo "🛑 توقف سرویس‌های قبلی..."
docker-compose down || true

echo "🧹 حذف کانتینرهای قدیمی..."
docker rm -f telegram-bot-postgres telegram-bot-redis 2>/dev/null || true

echo "🧺 حذف Volumeهای دیتابیس..."
docker volume rm telegram-bot-manager_postgres_data telegram-bot-manager_redis_data 2>/dev/null || true

echo "🚀 ساخت مجدد کانتینرها..."
docker-compose up -d

echo "⏳ چند ثانیه صبر برای راه‌اندازی Postgres..."
sleep 5

echo "✅ دیتابیس و ربات جدید با موفقیت بالا آمدند!"
docker ps

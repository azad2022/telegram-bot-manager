#!/bin/bash
set -e

# 🎨 رنگ‌ها برای خروجی ترمینال
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # بدون رنگ

echo -e "${BLUE}🛑 توقف سرویس‌های قبلی...${NC}"
docker-compose down || true

echo -e "${YELLOW}🧹 حذف کانتینرهای قدیمی...${NC}"
docker rm -f telegram-bot-postgres telegram-bot-redis 2>/dev/null || true

echo -e "${YELLOW}🧺 حذف Volumeهای دیتابیس...${NC}"
docker volume rm telegram-bot-manager_postgres_data telegram-bot-manager_redis_data 2>/dev/null || true

echo -e "${BLUE}🚀 ساخت مجدد کانتینرها...${NC}"
docker-compose up -d

echo -e "${YELLOW}⏳ چند ثانیه صبر برای راه‌اندازی Postgres و Redis...${NC}"
sleep 8

echo -e "${BLUE}🔍 بررسی وضعیت کانتینرها...${NC}"
docker ps

echo -e "${YELLOW}🧠 بررسی سلامت PostgreSQL...${NC}"
PG_CONTAINER=$(docker ps --filter "name=telegram-bot-postgres" --format "{{.ID}}")

if docker exec -it "$PG_CONTAINER" pg_isready -U bot_user > /dev/null 2>&1; then
  echo -e "${GREEN}✅ PostgreSQL با موفقیت بالا آمد!${NC}"
else
  echo -e "${RED}❌ خطا: PostgreSQL هنوز آماده نیست.${NC}"
  exit 1
fi

echo -e "${YELLOW}⚡ بررسی سلامت Redis...${NC}"
REDIS_CONTAINER=$(docker ps --filter "name=telegram-bot-redis" --format "{{.ID}}")

if docker exec -it "$REDIS_CONTAINER" redis-cli ping | grep -q "PONG"; then
  echo -e "${GREEN}✅ Redis با موفقیت بالا آمد!${NC}"
else
  echo -e "${RED}❌ خطا: Redis هنوز آماده نیست.${NC}"
  exit 1
fi

echo -e "${GREEN}🎉 عملیات با موفقیت انجام شد! دیتابیس و سرویس‌ها از صفر ساخته شدند.${NC}"

package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config — все настройки сервиса берутся из переменных окружения.
type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	TelegramToken   string        // токен бота из @BotFather
	TelegramBotName string        // username бота без @, для deep-link t.me/<name>
	CheckInterval   time.Duration // как часто планировщик проверяет цены
	WBDetailURL     string        // переопределение адреса WB API (если задан)
	TelegramProxy   string        // прокси для api.telegram.org (если заблокирован)
}

// Load читает .env (если есть) и собирает конфигурацию.
func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:            getenv("PORT", "8080"),
		DatabaseURL:     getenv("DATABASE_URL", "postgres://pricewatch:pricewatch@localhost:5432/pricewatch?sslmode=disable"),
		JWTSecret:       getenv("JWT_SECRET", "dev-secret-change-me"),
		TelegramToken:   getenv("TELEGRAM_BOT_TOKEN", ""),
		TelegramBotName: getenv("TELEGRAM_BOT_NAME", ""),
		CheckInterval:   time.Duration(getenvInt("CHECK_INTERVAL_MINUTES", 30)) * time.Minute,
		WBDetailURL:     getenv("WB_DETAIL_URL", ""),
		TelegramProxy:   getenv("TELEGRAM_PROXY", ""),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

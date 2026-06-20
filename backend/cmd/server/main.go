package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pricewatch/internal/api"
	"pricewatch/internal/auth"
	"pricewatch/internal/config"
	"pricewatch/internal/db"
	"pricewatch/internal/notify"
	"pricewatch/internal/parser"
	"pricewatch/internal/scheduler"
	"pricewatch/internal/store"
	"pricewatch/internal/telegrambot"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("БД: %v", err)
	}
	defer pool.Close()

	st := store.New(pool)
	authSvc := auth.New(cfg.JWTSecret)
	// Источник цен — Wildberries. При недоступности возвращается ошибка
	// (без подмены данными). Точный адрес API можно задать через WB_DETAIL_URL.
	wb := parser.NewWB()
	if cfg.WBDetailURL != "" {
		wb.UseEndpoints(cfg.WBDetailURL)
		log.Printf("WB endpoint переопределён: %s", cfg.WBDetailURL)
	}
	tg := notify.NewTelegram(cfg.TelegramToken, cfg.TelegramProxy)
	checker := scheduler.NewChecker(st, wb, tg)

	// фоновые процессы
	go checker.Run(ctx, cfg.CheckInterval)
	go telegrambot.New(tg, st).Run(ctx)

	srv := api.NewServer(st, authSvc, wb, checker, cfg.TelegramBotName)
	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: srv.Router(),
	}

	go func() {
		log.Printf("HTTP сервер запущен на :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP сервер: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("останавливаюсь...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

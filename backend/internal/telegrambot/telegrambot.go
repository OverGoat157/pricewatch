// Package telegrambot обрабатывает входящие сообщения бота — команду /start <код>
// для привязки Telegram-чата к аккаунту пользователя.
package telegrambot

import (
	"context"
	"log"
	"strings"
	"time"

	"pricewatch/internal/notify"
	"pricewatch/internal/store"
)

type Linker struct {
	tg    *notify.Telegram
	store *store.Store
}

func New(tg *notify.Telegram, st *store.Store) *Linker {
	return &Linker{tg: tg, store: st}
}

// Run запускает long-polling. Завершается при отмене контекста.
func (l *Linker) Run(ctx context.Context) {
	if !l.tg.Enabled() {
		log.Println("telegram: токен не задан — обработка привязки не запущена")
		return
	}
	log.Println("telegram: бот привязки запущен")

	var offset int64
	for {
		if ctx.Err() != nil {
			return
		}
		updates, err := l.tg.GetUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("telegram getUpdates: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		for _, u := range updates {
			offset = u.UpdateID + 1
			l.handle(ctx, u)
		}
	}
}

func (l *Linker) handle(ctx context.Context, u notify.Update) {
	if u.Message == nil {
		return
	}
	chatID := u.Message.Chat.ID
	text := strings.TrimSpace(u.Message.Text)

	if !strings.HasPrefix(text, "/start") {
		return
	}

	parts := strings.Fields(text)
	if len(parts) < 2 {
		_ = l.tg.SendMessage(ctx, chatID,
			"Привет! Чтобы получать уведомления, откройте «Настройки» в PriceWatch и нажмите «Подключить Telegram».")
		return
	}

	token := parts[1]
	user, err := l.store.GetUserByLinkToken(ctx, token)
	if err != nil {
		_ = l.tg.SendMessage(ctx, chatID,
			"❌ Код привязки не найден или уже использован. Сгенерируйте новый в настройках.")
		return
	}
	if err := l.store.LinkTelegram(ctx, user.ID, chatID); err != nil {
		log.Printf("telegram: привязка: %v", err)
		_ = l.tg.SendMessage(ctx, chatID, "Не удалось привязать аккаунт, попробуйте позже.")
		return
	}
	_ = l.tg.SendMessage(ctx, chatID,
		"✅ Аккаунт <b>"+user.Email+"</b> привязан.\nТеперь вы будете получать уведомления о снижении цен.")
}

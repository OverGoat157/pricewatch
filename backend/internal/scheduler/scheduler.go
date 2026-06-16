// Package scheduler — фоновая проверка цен и рассылка уведомлений.
package scheduler

import (
	"context"
	"fmt"
	"html"
	"log"
	"time"

	"pricewatch/internal/models"
	"pricewatch/internal/notify"
	"pricewatch/internal/parser"
	"pricewatch/internal/store"
)

type Checker struct {
	store    *store.Store
	parser   parser.Parser
	notifier *notify.Telegram
}

func NewChecker(st *store.Store, p parser.Parser, n *notify.Telegram) *Checker {
	return &Checker{store: st, parser: p, notifier: n}
}

// Run запускает периодическую проверку всех товаров.
func (c *Checker) Run(ctx context.Context, interval time.Duration) {
	log.Printf("планировщик: проверка цен каждые %s", interval)
	c.CheckAll(ctx) // первая проверка сразу при старте

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.CheckAll(ctx)
		}
	}
}

// CheckAll проверяет все товары в каталоге.
func (c *Checker) CheckAll(ctx context.Context) {
	products, err := c.store.ListProducts(ctx)
	if err != nil {
		log.Printf("планировщик: список товаров: %v", err)
		return
	}
	for _, p := range products {
		if err := c.CheckProduct(ctx, p); err != nil {
			log.Printf("планировщик: товар %d (%s): %v", p.ID, p.ExternalID, err)
		}
	}
}

// CheckProduct опрашивает один товар, пишет точку истории, обновляет последнюю
// цену и шлёт уведомления подписчикам, у которых цена пересекла цель сверху вниз.
func (c *Checker) CheckProduct(ctx context.Context, p models.Product) error {
	info, err := c.parser.Fetch(ctx, p.ExternalID)
	if err != nil {
		return err
	}

	prevPrice, hadPrev, err := c.store.PreviousPrice(ctx, p.ID)
	if err != nil {
		return err
	}
	if err := c.store.AddPricePoint(ctx, p.ID, info.Price, info.IsAvailable); err != nil {
		return err
	}
	if err := c.store.UpdateProductPrice(ctx, p.ID, info.Price, info.IsAvailable, info.Title, info.ImageURL); err != nil {
		return err
	}

	c.notifySubscribers(ctx, p, info, prevPrice, hadPrev)
	return nil
}

// notifySubscribers рассылает уведомления тем, для кого цена ТОЛЬКО ЧТО опустилась
// до целевой или ниже. Условие «пересечения» (предыдущая цена была выше цели)
// защищает от повторного спама, пока цена держится низкой.
func (c *Checker) notifySubscribers(ctx context.Context, p models.Product, info parser.ProductInfo, prevPrice int64, hadPrev bool) {
	subs, err := c.store.ActiveSubscriptionsForProduct(ctx, p.ID)
	if err != nil {
		log.Printf("планировщик: подписки товара %d: %v", p.ID, err)
		return
	}
	for _, sub := range subs {
		reached := info.Price <= sub.TargetPrice
		crossed := !hadPrev || prevPrice > sub.TargetPrice
		if reached && crossed {
			c.sendNotification(ctx, sub, p, info)
		}
	}
}

func (c *Checker) sendNotification(ctx context.Context, sub models.Subscription, p models.Product, info parser.ProductInfo) {
	title := info.Title
	if title == "" {
		title = p.Title
	}

	n := models.Notification{
		UserID:         sub.UserID,
		ProductID:      p.ID,
		TargetPrice:    sub.TargetPrice,
		TriggeredPrice: info.Price,
		Message: fmt.Sprintf("Цена снизилась: %s — %s (цель %s)",
			title, rub(info.Price), rub(sub.TargetPrice)),
	}
	if _, err := c.store.AddNotification(ctx, n); err != nil {
		log.Printf("планировщик: запись уведомления: %v", err)
	}

	user, err := c.store.GetUserByID(ctx, sub.UserID)
	if err != nil {
		log.Printf("планировщик: пользователь %d: %v", sub.UserID, err)
		return
	}
	if user.TelegramChatID == nil || !c.notifier.Enabled() {
		return
	}

	msg := fmt.Sprintf(
		"🔔 <b>Цена снизилась!</b>\n\n%s\nТекущая цена: <b>%s</b>\nВаша цель: %s\n\n<a href=\"%s\">Открыть товар</a>",
		html.EscapeString(title), rub(info.Price), rub(sub.TargetPrice), p.URL)
	if err := c.notifier.SendMessage(ctx, *user.TelegramChatID, msg); err != nil {
		log.Printf("планировщик: отправка в telegram: %v", err)
	}
}

// rub форматирует копейки в строку вида "2 899 ₽" / "1 234,50 ₽".
func rub(kopecks int64) string {
	whole := kopecks / 100
	cents := kopecks % 100
	s := thousands(whole)
	if cents == 0 {
		return s + " ₽"
	}
	return fmt.Sprintf("%s,%02d ₽", s, cents)
}

func thousands(n int64) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	var out []byte
	for i, ch := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, ch)
	}
	return string(out)
}

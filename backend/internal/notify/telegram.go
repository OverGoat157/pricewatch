// Package notify — отправка уведомлений через Telegram Bot API.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Telegram struct {
	token  string
	client *http.Client
}

// NewTelegram создаёт клиент Telegram. proxyURL (необязательно) задаёт прокси —
// http://, https:// или socks5:// — на случай, когда api.telegram.org недоступен
// напрямую (например, заблокирован у провайдера). WB-запросы идут отдельно и
// этот прокси не используют.
func NewTelegram(token, proxyURL string) *Telegram {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(u)
			log.Printf("telegram: запросы идут через прокси %s", u.Redacted())
		} else {
			log.Printf("telegram: некорректный TELEGRAM_PROXY (%v), прокси не используется", err)
		}
	}
	return &Telegram{
		token:  token,
		client: &http.Client{Timeout: 35 * time.Second, Transport: transport},
	}
}

// Enabled — задан ли токен бота.
func (t *Telegram) Enabled() bool { return t.token != "" }

func (t *Telegram) method(name string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.token, name)
}

// SendMessage отправляет сообщение в чат (HTML-разметка).
func (t *Telegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	if !t.Enabled() {
		return fmt.Errorf("telegram: токен бота не задан")
	}
	payload, _ := json.Marshal(map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.method("sendMessage"), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		return fmt.Errorf("telegram sendMessage: статус %d: %s", resp.StatusCode, b)
	}
	return nil
}

// Update — входящее обновление (нужны только сообщения с текстом).
type Update struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

// GetUpdates — long polling входящих сообщений (для привязки аккаунта).
func (t *Telegram) GetUpdates(ctx context.Context, offset int64) ([]Update, error) {
	if !t.Enabled() {
		return nil, fmt.Errorf("telegram: токен бота не задан")
	}
	reqURL := fmt.Sprintf("%s?offset=%d&timeout=25", t.method("getUpdates"), offset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if !out.OK {
		return nil, fmt.Errorf("telegram getUpdates: ok=false")
	}
	return out.Result, nil
}

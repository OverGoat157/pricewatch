package api

import (
	"math"
	"time"

	"pricewatch/internal/models"
)

type userDTO struct {
	ID             int64  `json:"id"`
	Email          string `json:"email"`
	TelegramLinked bool   `json:"telegram_linked"`
}

type productDTO struct {
	ID            int64      `json:"id"`
	Source        string     `json:"source"`
	ExternalID    string     `json:"external_id"`
	URL           string     `json:"url"`
	Title         string     `json:"title"`
	ImageURL      string     `json:"image_url"`
	LastPrice     *float64   `json:"last_price"` // рубли
	IsAvailable   bool       `json:"is_available"`
	LastCheckedAt *time.Time `json:"last_checked_at"`
}

type subscriptionDTO struct {
	ID          int64      `json:"id"`
	TargetPrice float64    `json:"target_price"` // рубли
	IsActive    bool       `json:"is_active"`
	BelowTarget bool       `json:"below_target"`
	CreatedAt   time.Time  `json:"created_at"`
	Product     productDTO `json:"product"`
}

type historyPointDTO struct {
	Price       float64   `json:"price"` // рубли
	IsAvailable bool      `json:"is_available"`
	CheckedAt   time.Time `json:"checked_at"`
}

type notificationDTO struct {
	ID             int64     `json:"id"`
	ProductID      int64     `json:"product_id"`
	ProductTitle   string    `json:"product_title"`
	TargetPrice    float64   `json:"target_price"`
	TriggeredPrice float64   `json:"triggered_price"`
	Message        string    `json:"message"`
	IsRead         bool      `json:"is_read"`
	SentAt         time.Time `json:"sent_at"`
}

func toUserDTO(u models.User) userDTO {
	return userDTO{ID: u.ID, Email: u.Email, TelegramLinked: u.TelegramChatID != nil}
}

func toProductDTO(p models.Product) productDTO {
	dto := productDTO{
		ID: p.ID, Source: p.Source, ExternalID: p.ExternalID, URL: p.URL,
		Title: p.Title, ImageURL: p.ImageURL, IsAvailable: p.IsAvailable,
		LastCheckedAt: p.LastCheckedAt,
	}
	if p.LastPrice != nil {
		r := kopToRub(*p.LastPrice)
		dto.LastPrice = &r
	}
	return dto
}

func toSubscriptionDTO(v models.SubscriptionView) subscriptionDTO {
	below := v.Product.LastPrice != nil && *v.Product.LastPrice <= v.TargetPrice
	return subscriptionDTO{
		ID:          v.ID,
		TargetPrice: kopToRub(v.TargetPrice),
		IsActive:    v.IsActive,
		BelowTarget: below,
		CreatedAt:   v.CreatedAt,
		Product:     toProductDTO(v.Product),
	}
}

func toNotificationDTO(n models.Notification) notificationDTO {
	return notificationDTO{
		ID: n.ID, ProductID: n.ProductID, ProductTitle: n.ProductTitle,
		TargetPrice: kopToRub(n.TargetPrice), TriggeredPrice: kopToRub(n.TriggeredPrice),
		Message: n.Message, IsRead: n.IsRead, SentAt: n.SentAt,
	}
}

func kopToRub(k int64) float64 { return float64(k) / 100 }
func rubToKop(r float64) int64 { return int64(math.Round(r * 100)) }

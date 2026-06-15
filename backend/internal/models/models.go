package models

import "time"

type User struct {
	ID             int64
	Email          string
	PasswordHash   string
	TelegramChatID *int64
	LinkToken      *string
	CreatedAt      time.Time
}

type Product struct {
	ID            int64
	Source        string
	ExternalID    string
	URL           string
	Title         string
	ImageURL      string
	LastPrice     *int64 // копейки
	IsAvailable   bool
	LastCheckedAt *time.Time
}

type Subscription struct {
	ID          int64
	UserID      int64
	ProductID   int64
	TargetPrice int64 // копейки
	IsActive    bool
	CreatedAt   time.Time
}

// SubscriptionView — подписка вместе с данными товара (для списка на дашборде).
type SubscriptionView struct {
	Subscription
	Product Product
}

type PricePoint struct {
	Price       int64 // копейки
	IsAvailable bool
	CheckedAt   time.Time
}

type Notification struct {
	ID             int64
	UserID         int64
	ProductID      int64
	ProductTitle   string
	TargetPrice    int64 // копейки
	TriggeredPrice int64 // копейки
	Message        string
	IsRead         bool
	SentAt         time.Time
}

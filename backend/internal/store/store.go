// Package store — слой доступа к данным (репозиторий) поверх pgx.
package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pricewatch/internal/models"
)

// ErrNotFound возвращается, когда запись не найдена.
var ErrNotFound = errors.New("не найдено")

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// ---------- Пользователи ----------

func (s *Store) CreateUser(ctx context.Context, email, hash string) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id, email, password_hash, telegram_chat_id, link_token, created_at`,
		email, hash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TelegramChatID, &u.LinkToken, &u.CreatedAt)
	return u, err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	return s.scanUser(s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, telegram_chat_id, link_token, created_at
		 FROM users WHERE email = $1`, email))
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	return s.scanUser(s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, telegram_chat_id, link_token, created_at
		 FROM users WHERE id = $1`, id))
}

func (s *Store) GetUserByLinkToken(ctx context.Context, token string) (models.User, error) {
	return s.scanUser(s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, telegram_chat_id, link_token, created_at
		 FROM users WHERE link_token = $1`, token))
}

func (s *Store) scanUser(row pgx.Row) (models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TelegramChatID, &u.LinkToken, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *Store) SetLinkToken(ctx context.Context, userID int64, token string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET link_token = $2 WHERE id = $1`, userID, token)
	return err
}

// LinkTelegram привязывает chat_id к пользователю и гасит одноразовый токен.
func (s *Store) LinkTelegram(ctx context.Context, userID, chatID int64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET telegram_chat_id = $2, link_token = NULL WHERE id = $1`, userID, chatID)
	return err
}

// ---------- Товары ----------

// UpsertProduct находит товар по (source, external_id) или создаёт новый.
// Существующие непустые title/image_url не затираются.
func (s *Store) UpsertProduct(ctx context.Context, source, externalID, url, title, imageURL string) (models.Product, error) {
	var p models.Product
	err := s.pool.QueryRow(ctx,
		`INSERT INTO products (source, external_id, url, title, image_url)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (source, external_id) DO UPDATE SET
		    url       = EXCLUDED.url,
		    title     = COALESCE(NULLIF(products.title, ''), EXCLUDED.title),
		    image_url = COALESCE(NULLIF(products.image_url, ''), EXCLUDED.image_url)
		 RETURNING id, source, external_id, url, title, image_url, last_price, is_available, last_checked_at`,
		source, externalID, url, title, imageURL,
	).Scan(&p.ID, &p.Source, &p.ExternalID, &p.URL, &p.Title, &p.ImageURL, &p.LastPrice, &p.IsAvailable, &p.LastCheckedAt)
	return p, err
}

func (s *Store) GetProduct(ctx context.Context, id int64) (models.Product, error) {
	var p models.Product
	err := s.pool.QueryRow(ctx,
		`SELECT id, source, external_id, url, title, image_url, last_price, is_available, last_checked_at
		 FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Source, &p.ExternalID, &p.URL, &p.Title, &p.ImageURL, &p.LastPrice, &p.IsAvailable, &p.LastCheckedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

// ListProducts — все товары (для планировщика).
func (s *Store) ListProducts(ctx context.Context) ([]models.Product, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, source, external_id, url, title, image_url, last_price, is_available, last_checked_at
		 FROM products ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Source, &p.ExternalID, &p.URL, &p.Title, &p.ImageURL,
			&p.LastPrice, &p.IsAvailable, &p.LastCheckedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpdateProductPrice(ctx context.Context, id, price int64, available bool, title, imageURL string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE products SET
		    last_price      = $2,
		    is_available    = $3,
		    last_checked_at = now(),
		    title           = CASE WHEN title = '' THEN $4 ELSE title END,
		    image_url       = CASE WHEN image_url = '' THEN $5 ELSE image_url END
		 WHERE id = $1`, id, price, available, title, imageURL)
	return err
}

// ---------- Подписки ----------

func (s *Store) CreateSubscription(ctx context.Context, userID, productID, target int64) (models.Subscription, error) {
	var sub models.Subscription
	err := s.pool.QueryRow(ctx,
		`INSERT INTO subscriptions (user_id, product_id, target_price)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, product_id) DO UPDATE SET target_price = EXCLUDED.target_price, is_active = true, notified = false
		 RETURNING id, user_id, product_id, target_price, is_active, created_at`,
		userID, productID, target,
	).Scan(&sub.ID, &sub.UserID, &sub.ProductID, &sub.TargetPrice, &sub.IsActive, &sub.CreatedAt)
	return sub, err
}

func (s *Store) ListUserSubscriptions(ctx context.Context, userID int64) ([]models.SubscriptionView, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT sub.id, sub.user_id, sub.product_id, sub.target_price, sub.is_active, sub.created_at,
		        p.id, p.source, p.external_id, p.url, p.title, p.image_url, p.last_price, p.is_available, p.last_checked_at
		 FROM subscriptions sub
		 JOIN products p ON p.id = sub.product_id
		 WHERE sub.user_id = $1
		 ORDER BY sub.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.SubscriptionView
	for rows.Next() {
		var v models.SubscriptionView
		if err := rows.Scan(
			&v.ID, &v.UserID, &v.ProductID, &v.TargetPrice, &v.IsActive, &v.CreatedAt,
			&v.Product.ID, &v.Product.Source, &v.Product.ExternalID, &v.Product.URL, &v.Product.Title,
			&v.Product.ImageURL, &v.Product.LastPrice, &v.Product.IsAvailable, &v.Product.LastCheckedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) GetSubscription(ctx context.Context, userID, subID int64) (models.SubscriptionView, error) {
	var v models.SubscriptionView
	err := s.pool.QueryRow(ctx,
		`SELECT sub.id, sub.user_id, sub.product_id, sub.target_price, sub.is_active, sub.created_at,
		        p.id, p.source, p.external_id, p.url, p.title, p.image_url, p.last_price, p.is_available, p.last_checked_at
		 FROM subscriptions sub
		 JOIN products p ON p.id = sub.product_id
		 WHERE sub.user_id = $1 AND sub.id = $2`, userID, subID,
	).Scan(
		&v.ID, &v.UserID, &v.ProductID, &v.TargetPrice, &v.IsActive, &v.CreatedAt,
		&v.Product.ID, &v.Product.Source, &v.Product.ExternalID, &v.Product.URL, &v.Product.Title,
		&v.Product.ImageURL, &v.Product.LastPrice, &v.Product.IsAvailable, &v.Product.LastCheckedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}

func (s *Store) UpdateSubscription(ctx context.Context, userID, subID, target int64, isActive bool) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE subscriptions SET target_price = $3, is_active = $4, notified = false WHERE user_id = $1 AND id = $2`,
		userID, subID, target, isActive)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteSubscription(ctx context.Context, userID, subID int64) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM subscriptions WHERE user_id = $1 AND id = $2`, userID, subID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ActiveSubscriptionsForProduct — активные подписки на товар (для рассылки уведомлений).
func (s *Store) ActiveSubscriptionsForProduct(ctx context.Context, productID int64) ([]models.Subscription, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, product_id, target_price, is_active, notified, created_at
		 FROM subscriptions WHERE product_id = $1 AND is_active = true`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.ProductID, &sub.TargetPrice, &sub.IsActive, &sub.Notified, &sub.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

// SetSubscriptionNotified помечает, отправлено ли уведомление о достижении цели.
func (s *Store) SetSubscriptionNotified(ctx context.Context, subID int64, notified bool) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE subscriptions SET notified = $2 WHERE id = $1`, subID, notified)
	return err
}

// ---------- История цен ----------

func (s *Store) AddPricePoint(ctx context.Context, productID, price int64, available bool) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO price_history (product_id, price, is_available) VALUES ($1, $2, $3)`,
		productID, price, available)
	return err
}

// PreviousPrice — последняя записанная цена товара (до добавления новой точки).
// found=false, если истории ещё нет.
func (s *Store) PreviousPrice(ctx context.Context, productID int64) (price int64, found bool, err error) {
	err = s.pool.QueryRow(ctx,
		`SELECT price FROM price_history WHERE product_id = $1 ORDER BY checked_at DESC LIMIT 1`,
		productID).Scan(&price)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return price, true, nil
}

func (s *Store) History(ctx context.Context, productID int64, limit int) ([]models.PricePoint, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT price, is_available, checked_at
		 FROM (
		    SELECT price, is_available, checked_at FROM price_history
		    WHERE product_id = $1 ORDER BY checked_at DESC LIMIT $2
		 ) t ORDER BY checked_at ASC`, productID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.PricePoint
	for rows.Next() {
		var pp models.PricePoint
		if err := rows.Scan(&pp.Price, &pp.IsAvailable, &pp.CheckedAt); err != nil {
			return nil, err
		}
		out = append(out, pp)
	}
	return out, rows.Err()
}

// ---------- Уведомления ----------

func (s *Store) AddNotification(ctx context.Context, n models.Notification) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx,
		`INSERT INTO notifications (user_id, product_id, target_price, triggered_price, message)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		n.UserID, n.ProductID, n.TargetPrice, n.TriggeredPrice, n.Message).Scan(&id)
	return id, err
}

func (s *Store) ListUserNotifications(ctx context.Context, userID int64, limit int) ([]models.Notification, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT n.id, n.user_id, n.product_id, p.title, n.target_price, n.triggered_price,
		        n.message, n.is_read, n.sent_at
		 FROM notifications n
		 JOIN products p ON p.id = n.product_id
		 WHERE n.user_id = $1
		 ORDER BY n.sent_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.ProductID, &n.ProductTitle, &n.TargetPrice,
			&n.TriggeredPrice, &n.Message, &n.IsRead, &n.SentAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`, userID)
	return err
}

func (s *Store) UnreadCount(ctx context.Context, userID int64) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userID).Scan(&n)
	return n, err
}

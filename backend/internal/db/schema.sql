-- Схема БД сервиса мониторинга цен.
-- Все денежные значения хранятся в КОПЕЙКАХ (BIGINT), чтобы не терять точность.

CREATE TABLE IF NOT EXISTS users (
    id               BIGSERIAL PRIMARY KEY,
    email            TEXT        NOT NULL UNIQUE,
    password_hash    TEXT        NOT NULL,
    telegram_chat_id BIGINT,                 -- chat_id в Telegram (после привязки)
    link_token       TEXT,                   -- одноразовый код для привязки бота
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Общий каталог отслеживаемых товаров (один товар = одна запись, даже если его
-- отслеживают несколько пользователей). Это избавляет от повторного парсинга.
CREATE TABLE IF NOT EXISTS products (
    id              BIGSERIAL PRIMARY KEY,
    source          TEXT        NOT NULL,            -- 'wildberries'
    external_id     TEXT        NOT NULL,            -- артикул (nmId)
    url             TEXT        NOT NULL,
    title           TEXT        NOT NULL DEFAULT '',
    image_url       TEXT        NOT NULL DEFAULT '',
    last_price      BIGINT,                          -- последняя известная цена, копейки
    is_available    BOOLEAN     NOT NULL DEFAULT true,
    last_checked_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source, external_id)
);

-- Подписка пользователя на товар с целевой ценой.
CREATE TABLE IF NOT EXISTS subscriptions (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT      NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    product_id   BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    target_price BIGINT      NOT NULL,                        -- желаемая цена, копейки
    is_active    BOOLEAN     NOT NULL DEFAULT true,
    notified     BOOLEAN     NOT NULL DEFAULT false,          -- уведомление о достижении цели уже отправлено
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, product_id)
);
-- для уже существующих БД (таблица создана без notified)
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS notified BOOLEAN NOT NULL DEFAULT false;

-- История цен по товару (точки графика).
CREATE TABLE IF NOT EXISTS price_history (
    id           BIGSERIAL PRIMARY KEY,
    product_id   BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price        BIGINT      NOT NULL,
    is_available BOOLEAN     NOT NULL DEFAULT true,
    checked_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_price_history_product ON price_history (product_id, checked_at);

-- Журнал отправленных уведомлений (для колокольчика в UI и истории).
CREATE TABLE IF NOT EXISTS notifications (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT      NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    product_id      BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    target_price    BIGINT      NOT NULL,
    triggered_price BIGINT      NOT NULL,
    message         TEXT        NOT NULL DEFAULT '',
    is_read         BOOLEAN     NOT NULL DEFAULT false,
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications (user_id, sent_at DESC);

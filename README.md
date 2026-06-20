# PriceWatch — сервис мониторинга цен на товары

Веб-приложение для отслеживания цен товаров на Wildberries: парсинг текущей
цены, история изменений в виде графика и уведомления в Telegram, когда цена
опускается до заданной пользователем.

Итоговая аттестационная работа по программе ДПО «Фронтенд и бэкенд разработка».

## Возможности

- 🔐 Регистрация и вход (JWT, пароли хешируются bcrypt), у каждого пользователя свой список товаров.
- ➕ Добавление товара по ссылке или артикулу Wildberries с указанием целевой цены.
- 🤖 Фоновый парсинг цен по расписанию (планировщик) + кнопка «Проверить сейчас».
- 📈 История цен: график (Chart.js) и таблица проверок, минимум/максимум/текущая.
- 📨 Уведомления в Telegram при снижении цены до целевой + журнал уведомлений в интерфейсе.

## Стек

| Слой        | Технологии                                            |
|-------------|-------------------------------------------------------|
| Backend     | Go, chi (HTTP), pgx (PostgreSQL), JWT, bcrypt         |
| Frontend    | Vue 3 (Composition API), Vue Router, Pinia, Axios, Chart.js, Vite |
| База данных | PostgreSQL                                            |
| Инфра       | Docker, Docker Compose, nginx                         |
| Внешние API | Wildberries (публичный JSON), Telegram Bot API        |

## Архитектура

```
Браузер ──▶ nginx (frontend) ──/api──▶ Go backend ──▶ PostgreSQL
                                          │
                                          ├─ Scheduler  — периодически парсит цены
                                          ├─ Parser     — Wildberries JSON API
                                          └─ Telegram   — long polling (привязка) + отправка уведомлений
```

## Быстрый запуск (Docker)

```bash
cp .env.example .env          # при желании впишите токен бота и порт APP_PORT
docker compose up --build
```

Откройте **http://localhost:8080** (порт меняется через `APP_PORT` в `.env`).
БД и backend наружу не публикуются — браузер ходит на фронтенд (nginx), а тот
проксирует `/api` на backend внутри docker-сети.

> Без `TELEGRAM_BOT_TOKEN` приложение полностью работает, но уведомления приходят
> только в журнал внутри сайта (без сообщений в Telegram).

## Локальный запуск (без Docker)

Нужны Go 1.24+, Node 18+, PostgreSQL.

```bash
# 1. БД
createdb pricewatch   # либо docker run -e POSTGRES_PASSWORD=pricewatch -e POSTGRES_USER=pricewatch -e POSTGRES_DB=pricewatch -p 5432:5432 postgres:16

# 2. Backend
cd backend
cp .env.example .env
go run ./cmd/server

# 3. Frontend (в отдельном терминале)
cd frontend
npm install
npm run dev      # http://localhost:5173, /api проксируется на :8080
```

## Настройка Telegram-бота

1. В Telegram напишите [@BotFather](https://t.me/BotFather) → `/newbot`, задайте имя.
2. Скопируйте **токен** и **username** бота.
3. Впишите их в `.env`:
   ```
   TELEGRAM_BOT_TOKEN=123456:ABC...
   TELEGRAM_BOT_NAME=my_pricewatch_bot
   ```
4. Перезапустите backend. В приложении: **Настройки → Подключить Telegram**, нажмите Start у бота, вернитесь и нажмите «Обновить статус».

## Проверка парсера Wildberries

Парсер обращается к публичному JSON API карточки товара. Проверить его доступность
из вашей сети можно так (подставьте реальный артикул):

```bash
curl "https://card.wb.ru/cards/v2/detail?appType=1&curr=rub&dest=-1257786&nm=179978204"
```

Используются только реальные данные Wildberries. Площадка периодически меняет
версию пути карточного API, поэтому парсер по очереди пробует несколько адресов
(`/cards/v2/detail`, `/cards/v1/detail`, `/cards/v3/detail`, `/cards/v4/detail`,
`/cards/detail`) и использует первый ответивший. Если ни один не доступен —
возвращается ошибка (502) и пишется лог; подменять реальные цены сгенерированными
данными приложение не будет.

Если известен точный рабочий адрес API, его можно задать без пересборки переменной
окружения `WB_DETAIL_URL` (тогда используется только он).

Логика разбора покрыта юнит-тестом на сохранённом ответе:

```bash
cd backend && go test ./...
```

## Наполнение истории для демонстрации

Чтобы график был наглядным на видео, можно сгенерировать историю цен для уже
добавленного товара (замените `1` на нужный `product_id`):

```sql
INSERT INTO price_history (product_id, price, is_available, checked_at)
SELECT 1,
       250000 + (random() * 80000)::bigint,  -- цена в копейках
       true,
       now() - (g || ' days')::interval
FROM generate_series(30, 1, -1) AS g;
```

## Структура проекта

```
pricewatch/
├── backend/
│   ├── cmd/server/main.go         # точка входа, сборка зависимостей
│   └── internal/
│       ├── api/                   # HTTP-слой: роутер, middleware, обработчики
│       ├── auth/                  # JWT + bcrypt
│       ├── config/                # конфигурация из ENV
│       ├── db/                    # подключение к БД, schema.sql
│       ├── models/                # доменные модели
│       ├── notify/                # Telegram Bot API
│       ├── parser/                # интерфейс Parser + реализация Wildberries (+тест)
│       ├── scheduler/             # фоновая проверка цен и рассылка уведомлений
│       ├── store/                 # слой доступа к данным (репозиторий)
│       └── telegrambot/           # обработка /start для привязки чата
├── frontend/
│   └── src/
│       ├── views/                 # страницы (вход, дашборд, товар, настройки)
│       ├── components/            # шапка, карточка товара, график, колокольчик
│       ├── stores/                # Pinia (авторизация)
│       ├── api.js, router.js, format.js
│       └── styles.css
├── docker-compose.yml
└── README.md
```

## Основные API-эндпоинты

| Метод  | Путь                              | Назначение                       |
|--------|-----------------------------------|----------------------------------|
| POST   | `/api/auth/register`              | регистрация                      |
| POST   | `/api/auth/login`                 | вход                             |
| GET    | `/api/me`                         | текущий пользователь             |
| GET    | `/api/subscriptions`              | список отслеживаемых товаров      |
| POST   | `/api/subscriptions`              | добавить товар                   |
| GET    | `/api/subscriptions/{id}`         | товар + история цен              |
| PATCH  | `/api/subscriptions/{id}`         | изменить цель / активность        |
| DELETE | `/api/subscriptions/{id}`         | удалить                          |
| POST   | `/api/subscriptions/{id}/check`   | проверить цену сейчас            |
| GET    | `/api/notifications`              | уведомления + счётчик непрочит.  |
| POST   | `/api/notifications/read`         | отметить все прочитанными        |
| GET    | `/api/telegram/link`              | код/ссылка для привязки бота      |

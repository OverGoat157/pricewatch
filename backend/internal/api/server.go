// Package api — HTTP-слой: маршруты, middleware, обработчики.
package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"pricewatch/internal/auth"
	"pricewatch/internal/parser"
	"pricewatch/internal/scheduler"
	"pricewatch/internal/store"
)

type Server struct {
	store   *store.Store
	auth    *auth.Service
	parser  parser.Parser
	checker *scheduler.Checker
	botName string
}

func NewServer(st *store.Store, a *auth.Service, p parser.Parser, ch *scheduler.Checker, botName string) *Server {
	return &Server{store: st, auth: a, parser: p, checker: ch, botName: botName}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Post("/api/auth/register", s.handleRegister)
	r.Post("/api/auth/login", s.handleLogin)

	// защищённые маршруты
	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)

		r.Get("/api/me", s.handleMe)

		r.Get("/api/subscriptions", s.handleListSubscriptions)
		r.Post("/api/subscriptions", s.handleCreateSubscription)
		r.Get("/api/subscriptions/{id}", s.handleGetSubscription)
		r.Patch("/api/subscriptions/{id}", s.handleUpdateSubscription)
		r.Delete("/api/subscriptions/{id}", s.handleDeleteSubscription)
		r.Post("/api/subscriptions/{id}/check", s.handleCheckSubscription)

		r.Get("/api/notifications", s.handleListNotifications)
		r.Post("/api/notifications/read", s.handleMarkAllRead)

		r.Get("/api/telegram/link", s.handleTelegramLink)
	})

	return r
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, prefix) {
			writeError(w, http.StatusUnauthorized, "требуется авторизация")
			return
		}
		uid, err := s.auth.ParseToken(strings.TrimPrefix(h, prefix))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "невалидный токен")
			return
		}
		next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), uid)))
	})
}

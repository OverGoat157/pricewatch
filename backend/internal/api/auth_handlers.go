package api

import (
	"errors"
	"net/http"
	"strings"

	"pricewatch/internal/auth"
	"pricewatch/internal/store"
)

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := decode(r, &c); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный запрос")
		return
	}
	c.Email = strings.TrimSpace(strings.ToLower(c.Email))
	if !strings.Contains(c.Email, "@") {
		writeError(w, http.StatusBadRequest, "введите корректный email")
		return
	}
	if len(c.Password) < 6 {
		writeError(w, http.StatusBadRequest, "пароль должен быть не короче 6 символов")
		return
	}

	if _, err := s.store.GetUserByEmail(r.Context(), c.Email); err == nil {
		writeError(w, http.StatusConflict, "пользователь с таким email уже существует")
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "ошибка сервера")
		return
	}

	hash, err := auth.HashPassword(c.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка сервера")
		return
	}
	u, err := s.store.CreateUser(r.Context(), c.Email, hash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось создать пользователя")
		return
	}
	s.issueToken(w, u.ID, toUserDTO(u))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := decode(r, &c); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный запрос")
		return
	}
	c.Email = strings.TrimSpace(strings.ToLower(c.Email))

	u, err := s.store.GetUserByEmail(r.Context(), c.Email)
	if err != nil || !auth.CheckPassword(u.PasswordHash, c.Password) {
		writeError(w, http.StatusUnauthorized, "неверный email или пароль")
		return
	}
	s.issueToken(w, u.ID, toUserDTO(u))
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, err := s.store.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "пользователь не найден")
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(u))
}

func (s *Server) issueToken(w http.ResponseWriter, uid int64, u userDTO) {
	token, err := s.auth.GenerateToken(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось создать токен")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: u})
}

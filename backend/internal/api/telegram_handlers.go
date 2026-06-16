package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func (s *Server) handleTelegramLink(w http.ResponseWriter, r *http.Request) {
	u, err := s.store.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "пользователь не найден")
		return
	}

	token, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка сервера")
		return
	}
	if err := s.store.SetLinkToken(r.Context(), u.ID, token); err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось подготовить привязку")
		return
	}

	resp := map[string]any{
		"token":    token,
		"bot_name": s.botName,
		"linked":   u.TelegramChatID != nil,
	}
	if s.botName != "" {
		resp["link"] = "https://t.me/" + s.botName + "?start=" + token
	}
	writeJSON(w, http.StatusOK, resp)
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

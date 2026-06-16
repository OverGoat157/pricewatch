package api

import "net/http"

func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListUserNotifications(r.Context(), userID(r), 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось получить уведомления")
		return
	}
	unread, _ := s.store.UnreadCount(r.Context(), userID(r))

	out := make([]notificationDTO, 0, len(items))
	for _, n := range items {
		out = append(out, toNotificationDTO(n))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out, "unread": unread})
}

func (s *Server) handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	if err := s.store.MarkAllNotificationsRead(r.Context(), userID(r)); err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось обновить уведомления")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"pricewatch/internal/store"
)

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func (s *Server) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	views, err := s.store.ListUserSubscriptions(r.Context(), userID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось получить список")
		return
	}
	out := make([]subscriptionDTO, 0, len(views))
	for _, v := range views {
		out = append(out, toSubscriptionDTO(v))
	}
	writeJSON(w, http.StatusOK, out)
}

type createSubReq struct {
	URL         string  `json:"url"`
	TargetPrice float64 `json:"target_price"`
}

func (s *Server) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req createSubReq
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный запрос")
		return
	}
	if req.TargetPrice <= 0 {
		writeError(w, http.StatusBadRequest, "укажите целевую цену больше нуля")
		return
	}

	externalID, ok := s.parser.ExternalID(req.URL)
	if !ok {
		writeError(w, http.StatusBadRequest, "не удалось распознать ссылку или артикул Wildberries")
		return
	}

	info, err := s.parser.Fetch(r.Context(), externalID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "не удалось получить данные товара (проверьте ссылку или повторите позже)")
		return
	}

	url := canonicalURL(req.URL, externalID)
	product, err := s.store.UpsertProduct(r.Context(), s.parser.Name(), externalID, url, info.Title, info.ImageURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось сохранить товар")
		return
	}

	_ = s.store.AddPricePoint(r.Context(), product.ID, info.Price, info.IsAvailable)
	_ = s.store.UpdateProductPrice(r.Context(), product.ID, info.Price, info.IsAvailable, info.Title, info.ImageURL)

	sub, err := s.store.CreateSubscription(r.Context(), userID(r), product.ID, rubToKop(req.TargetPrice))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось создать подписку")
		return
	}

	view, err := s.store.GetSubscription(r.Context(), userID(r), sub.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка сервера")
		return
	}
	writeJSON(w, http.StatusCreated, toSubscriptionDTO(view))
}

func (s *Server) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный id")
		return
	}
	view, err := s.store.GetSubscription(r.Context(), userID(r), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "подписка не найдена")
		return
	}

	points, err := s.store.History(r.Context(), view.ProductID, 500)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось получить историю")
		return
	}
	history := make([]historyPointDTO, 0, len(points))
	for _, p := range points {
		history = append(history, historyPointDTO{
			Price: kopToRub(p.Price), IsAvailable: p.IsAvailable, CheckedAt: p.CheckedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"subscription": toSubscriptionDTO(view),
		"history":      history,
	})
}

type updateSubReq struct {
	TargetPrice *float64 `json:"target_price"`
	IsActive    *bool    `json:"is_active"`
}

func (s *Server) handleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный id")
		return
	}
	view, err := s.store.GetSubscription(r.Context(), userID(r), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "подписка не найдена")
		return
	}

	var req updateSubReq
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный запрос")
		return
	}

	target := view.TargetPrice
	if req.TargetPrice != nil {
		if *req.TargetPrice <= 0 {
			writeError(w, http.StatusBadRequest, "целевая цена должна быть больше нуля")
			return
		}
		target = rubToKop(*req.TargetPrice)
	}
	isActive := view.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	if err := s.store.UpdateSubscription(r.Context(), userID(r), id, target, isActive); err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось обновить подписку")
		return
	}

	updated, _ := s.store.GetSubscription(r.Context(), userID(r), id)
	writeJSON(w, http.StatusOK, toSubscriptionDTO(updated))
}

func (s *Server) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный id")
		return
	}
	if err := s.store.DeleteSubscription(r.Context(), userID(r), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "подписка не найдена")
			return
		}
		writeError(w, http.StatusInternalServerError, "не удалось удалить подписку")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleCheckSubscription — «Проверить сейчас»: внеплановый опрос цены товара.
func (s *Server) handleCheckSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный id")
		return
	}
	view, err := s.store.GetSubscription(r.Context(), userID(r), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "подписка не найдена")
		return
	}
	product, err := s.store.GetProduct(r.Context(), view.ProductID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка сервера")
		return
	}
	if err := s.checker.CheckProduct(r.Context(), product); err != nil {
		writeError(w, http.StatusBadGateway, "не удалось получить актуальную цену")
		return
	}

	updated, _ := s.store.GetSubscription(r.Context(), userID(r), id)
	writeJSON(w, http.StatusOK, toSubscriptionDTO(updated))
}

func canonicalURL(input, externalID string) string {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "http") {
		return input
	}
	return "https://www.wildberries.ru/catalog/" + externalID + "/detail.aspx"
}

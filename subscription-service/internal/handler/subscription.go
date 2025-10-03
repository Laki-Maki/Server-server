package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"subscription-service/internal/model"
	"subscription-service/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Handler struct {
	svc service.SubscriptionService
	log *zerolog.Logger
}

func NewHandler(svc service.SubscriptionService, log *zerolog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func NewRouter(h *Handler, log *zerolog.Logger) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/subscriptions", h.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions", h.ListSubscriptions).Methods("GET")
	r.HandleFunc("/subscriptions/aggregate", h.Aggregate).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", h.GetSubscription).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", h.UpdateSubscription).Methods("PUT")
	r.HandleFunc("/subscriptions/{id}", h.DeleteSubscription).Methods("DELETE")

	return r
}

func parseMonthYear(param string) (time.Time, error) {
	t, err := time.Parse("01-2006", param)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ServiceName string  `json:"service_name"`
		Price       int     `json:"price"`
		UserID      string  `json:"user_id"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	// basic validation
	if strings.TrimSpace(in.ServiceName) == "" {
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}
	if in.Price < 0 {
		http.Error(w, "price must be >= 0", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(in.UserID); err != nil {
		http.Error(w, "user_id must be a valid UUID", http.StatusBadRequest)
		return
	}

	start, err := parseMonthYear(in.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format, expected MM-YYYY", http.StatusBadRequest)
		return
	}

	var end *time.Time
	if in.EndDate != nil && *in.EndDate != "" {
		t, err := parseMonthYear(*in.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
		if t.Before(start) {
			http.Error(w, "end_date must be the same or after start_date", http.StatusBadRequest)
			return
		}
		end = &t
	}

	sub := &model.Subscription{
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartDate:   start,
		EndDate:     end,
	}

	if err := h.svc.Create(r.Context(), sub); err != nil {
		h.log.Error().Err(err).Msg("create subscription failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// set Location and return 201
	w.Header().Set("Location", fmt.Sprintf("/subscriptions/%s", sub.ID))
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, sub)
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.log.Error().Err(err).Msg("get subscription failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, sub)
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := q.Get("user_id")
	serviceName := q.Get("service_name")

	limit := 50
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	offset := 0
	if o := q.Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	subs, err := h.svc.List(r.Context(), userID, serviceName, limit, offset)
	if err != nil {
		h.log.Error().Err(err).Msg("list subscriptions failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, subs)
}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var in struct {
		ServiceName string  `json:"service_name"`
		Price       int     `json:"price"`
		UserID      string  `json:"user_id"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	// basic validation
	if strings.TrimSpace(in.ServiceName) == "" {
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}
	if in.Price < 0 {
		http.Error(w, "price must be >= 0", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(in.UserID); err != nil {
		http.Error(w, "user_id must be a valid UUID", http.StatusBadRequest)
		return
	}

	start, err := parseMonthYear(in.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format, expected MM-YYYY", http.StatusBadRequest)
		return
	}

	var end *time.Time
	if in.EndDate != nil && *in.EndDate != "" {
		t, err := parseMonthYear(*in.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
		if t.Before(start) {
			http.Error(w, "end_date must be the same or after start_date", http.StatusBadRequest)
			return
		}
		end = &t
	}

	sub := &model.Subscription{
		ID:          id,
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartDate:   start,
		EndDate:     end,
	}

	if err := h.svc.Update(r.Context(), sub); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.log.Error().Err(err).Msg("update subscription failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// return the fresh record from DB (with timestamps)
	updated, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.log.Error().Err(err).Msg("fetch after update failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, updated)
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.log.Error().Err(err).Msg("delete subscription failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Aggregate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	if from == "" || to == "" {
		http.Error(w, "from and to are required (MM-YYYY)", http.StatusBadRequest)
		return
	}

	// validate format
	fromDate, err := parseMonthYear(from)
	if err != nil {
		http.Error(w, "invalid from format, expected MM-YYYY", http.StatusBadRequest)
		return
	}
	toDate, err := parseMonthYear(to)
	if err != nil {
		http.Error(w, "invalid to format, expected MM-YYYY", http.StatusBadRequest)
		return
	}

	// проверка from <= to
	if fromDate.After(toDate) {
		http.Error(w, "`from` must be less than or equal to `to`", http.StatusBadRequest)
		return
	}

	var userID *string
	if v := q.Get("user_id"); v != "" {
		userID = &v
	}
	var serviceName *string
	if v := q.Get("service_name"); v != "" {
		serviceName = &v
	}

	subs, total, err := h.svc.AggregateWithDetails(r.Context(), from, to, userID, serviceName)
	if err != nil {
		h.log.Error().Err(err).Msg("aggregate failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := model.AggregateResponse{
		From:          from,
		To:            to,
		Total:         total,
		Subscriptions: subs,
	}

	if userID != nil {
		response.UserID = *userID
	}

	writeJSON(w, response)
}

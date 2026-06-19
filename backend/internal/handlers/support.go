package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eco-rental/internal/database"
	"eco-rental/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTicket обробляє запит на створення нового звернення в техпідтримку
func CreateTicket(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDRaw := r.Context().Value(middleware.UserIDKey)
		userID, ok := userIDRaw.(int)
		if !ok {
			RespondWithError(w, http.StatusUnauthorized, "Помилка авторизації")
			return
		}

		var req struct {
			Subject string `json:"subject"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Невалідний JSON")
			return
		}
		if req.Subject == "" || req.Message == "" {
			RespondWithError(w, http.StatusBadRequest, "Тема і повідомлення не можуть бути порожніми")
			return
		}

		ticketID, err := database.CreateTicket(pool, userID, req.Subject, req.Message)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   "Звернення успішно створено",
			"ticket_id": ticketID,
		})
	}
}

// GetUserTickets повертає всі тікети авторизованого користувача
func GetUserTickets(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDRaw := r.Context().Value(middleware.UserIDKey)
		userID, ok := userIDRaw.(int)
		if !ok {
			RespondWithError(w, http.StatusUnauthorized, "Помилка авторизації")
			return
		}

		tickets, err := database.GetUserTickets(pool, userID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tickets)
	}
}

// GetAllTicketsAdmin повертає всі тікети системи для адміна
func GetAllTicketsAdmin(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tickets, err := database.GetAllTicketsAdmin(pool)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tickets)
	}
}

// UpdateTicketStatus оновлює статус тікета (наприклад, адмін закриває його)
func UpdateTicketStatus(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Некоректний ID тікета")
			return
		}

		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Невалідний JSON")
			return
		}

		if err := database.UpdateTicketStatus(pool, id, req.Status); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Заглушки для можливого розширення чату в майбутньому
func GetTicketDetails(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }
}

// ReplyToTicket обробляє додавання нового повідомлення в існуючий тікет
func ReplyToTicket(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDRaw := r.Context().Value(middleware.UserIDKey)
		userID, ok := userIDRaw.(int)
		if !ok {
			RespondWithError(w, http.StatusUnauthorized, "Помилка авторизації")
			return
		}

		ticketID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Некоректний ID тікета")
			return
		}

		var req struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
			RespondWithError(w, http.StatusBadRequest, "Повідомлення не може бути порожнім")
			return
		}

		if err := database.AddMessageToTicket(pool, ticketID, userID, req.Message); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

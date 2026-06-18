package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eco-rental/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetAllPromoCodesAdmin повертає список всіх промокодів
func GetAllPromoCodesAdmin(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promos, err := database.GetAllPromoCodesAdmin(pool)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(promos)
	}
}

// AddPromoCodeAdmin створює новий промокод
func AddPromoCodeAdmin(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Code            string  `json:"code"`
			Type            string  `json:"type"`
			RewardAmount    float64 `json:"reward_amount"`
			DiscountPercent int     `json:"discount_percent"`
			MaxUses         int     `json:"max_uses"`
			UserID          *int    `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Невірний формат запиту")
			return
		}

		id, err := database.AddPromoCodeAdmin(pool, req.Code, req.Type, req.RewardAmount, req.DiscountPercent, req.MaxUses, req.UserID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "message": "Промокод успішно створено"})
	}
}

// DeletePromoCodeAdmin видаляє промокод
func DeletePromoCodeAdmin(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil || id <= 0 {
			RespondWithError(w, http.StatusBadRequest, "Невірний ID")
			return
		}
		if err := database.DeletePromoCodeAdmin(pool, id); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Промокод видалено"})
	}
}

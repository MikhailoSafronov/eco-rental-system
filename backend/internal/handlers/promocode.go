package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"eco-rental/internal/database"
	"eco-rental/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplyPromoCode обробляє POST запит на застосування промокоду
func ApplyPromoCode(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDRaw := r.Context().Value(middleware.UserIDKey)
		userID, ok := userIDRaw.(int)
		if !ok {
			RespondWithError(w, http.StatusUnauthorized, "Помилка авторизації")
			return
		}

		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
			RespondWithError(w, http.StatusBadRequest, "Будь ласка, вкажіть промокод")
			return
		}

		reward, promoType, err := database.ApplyPromoCode(pool, userID, req.Code)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var msg string
		if promoType == "discount" {
			msg = fmt.Sprintf("Промокод успішно застосовано! Ви отримали знижку %d%% на наступну поїздку 🛴", int(reward))
		} else {
			msg = fmt.Sprintf("Промокод успішно застосовано! Вам нараховано %.2f ₴ 💸", reward)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": msg,
			"reward":  reward,
		})
	}
}

package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplyPromoCode застосовує промокод для користувача в єдиній транзакції
// Повертає значення бонусу, тип промокоду та помилку
func ApplyPromoCode(pool *pgxpool.Pool, userID int, code string) (float64, string, error) {
	ctx := context.Background()

	// Починаємо транзакцію
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, "", fmt.Errorf("помилка старту транзакції: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Шукаємо промокод і перевіряємо його актуальність
	var promoID, currentUses, maxUses, discountPercent int
	var rewardAmount float64
	var promoType string
	var targetUserID *int // Вказівник, оскільки поле в БД може бути NULL

	err = tx.QueryRow(ctx, `
		SELECT id, type, reward_amount, discount_percent, current_uses, max_uses, user_id
		FROM promo_codes 
		WHERE code = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`, code).Scan(&promoID, &promoType, &rewardAmount, &discountPercent, &currentUses, &maxUses, &targetUserID)

	if err != nil {
		return 0, "", errors.New("промокод не знайдено або його термін дії минув")
	}

	// Перевіряємо, чи це персональний промокод, і чи збігається ID
	if targetUserID != nil && *targetUserID != userID {
		return 0, "", errors.New("промокод не знайдено або він призначений для іншого акаунта")
	}

	// 2. Перевіряємо глобальний ліміт використань
	if currentUses >= maxUses {
		return 0, "", errors.New("на жаль, ліміт використання цього промокоду вичерпано")
	}

	// 3. Перевіряємо, чи цей користувач вже використовував цей промокод
	var alreadyUsed bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_promo_usages WHERE user_id = $1 AND promo_id = $2)
	`, userID, promoID).Scan(&alreadyUsed)

	if err != nil {
		return 0, "", fmt.Errorf("помилка перевірки історії: %w", err)
	}
	if alreadyUsed {
		return 0, "", errors.New("ви вже використовували цей промокод раніше")
	}

	// Якщо це знижка - перевіримо, чи немає вже активної
	if promoType == "discount" {
		var currentDiscount int
		tx.QueryRow(ctx, "SELECT active_discount_percent FROM users WHERE id = $1", userID).Scan(&currentDiscount)
		if currentDiscount > 0 {
			return 0, "", errors.New("у вас вже є активна знижка на наступну поїздку")
		}
	}

	// 4. Збільшуємо лічильник використання промокоду
	_, err = tx.Exec(ctx, "UPDATE promo_codes SET current_uses = current_uses + 1 WHERE id = $1", promoID)
	if err != nil {
		return 0, "", fmt.Errorf("помилка оновлення лічильника: %w", err)
	}

	// 5. Записуємо, що юзер успішно його застосував
	_, err = tx.Exec(ctx, "INSERT INTO user_promo_usages (user_id, promo_id) VALUES ($1, $2)", userID, promoID)
	if err != nil {
		return 0, "", fmt.Errorf("помилка запису використання: %w", err)
	}

	// 6. Застосовуємо бонус в залежності від типу
	if promoType == "discount" {
		_, err = tx.Exec(ctx, "UPDATE users SET active_discount_percent = $1 WHERE id = $2", discountPercent, userID)
		if err != nil {
			return 0, "", fmt.Errorf("помилка нарахування знижки: %w", err)
		}
		rewardAmount = float64(discountPercent) // Для відповіді клієнту
	} else {
		// Поповнюємо баланс
		_, err = tx.Exec(ctx, "UPDATE users SET balance = balance + $1 WHERE id = $2", rewardAmount, userID)
		if err != nil {
			return 0, "", fmt.Errorf("помилка нарахування балансу: %w", err)
		}

		// Робимо запис в історію транзакцій (тільки для грошових бонусів)
		// Додаємо userID до ID транзакції, щоб вона була гарантовано унікальною в payments
		transactionID := fmt.Sprintf("PROMO-%s-%d", code, userID)
		_, err = tx.Exec(ctx, `
			INSERT INTO payments (user_id, amount, type, status, external_transaction_id) 
			VALUES ($1, $2, 'top_up', 'succeeded', $3)
		`, userID, rewardAmount, transactionID)
		if err != nil {
			return 0, "", fmt.Errorf("помилка запису в історію платежів: %w", err)
		}
	}

	// 8. Якщо все добре — комітимо транзакцію
	if err = tx.Commit(ctx); err != nil {
		return 0, "", fmt.Errorf("помилка підтвердження транзакції: %w", err)
	}

	return rewardAmount, promoType, nil
}

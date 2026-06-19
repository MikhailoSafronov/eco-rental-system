package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTicket створює нове звернення та записує перше повідомлення
func CreateTicket(pool *pgxpool.Pool, userID int, subject, message string) (int, error) {
	ctx := context.Background()

	// Починаємо транзакцію
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("помилка старту транзакції: %w", err)
	}
	defer tx.Rollback(ctx)

	var ticketID int
	err = tx.QueryRow(ctx, "INSERT INTO support_tickets (user_id, subject) VALUES ($1, $2) RETURNING id", userID, subject).Scan(&ticketID)
	if err != nil {
		return 0, fmt.Errorf("помилка створення тікета: %w", err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO support_messages (ticket_id, sender_id, message) VALUES ($1, $2, $3)", ticketID, userID, message)
	if err != nil {
		return 0, fmt.Errorf("помилка збереження повідомлення: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("помилка збереження транзакції: %w", err)
	}

	return ticketID, nil
}

// GetUserTickets повертає всі тікети конкретного користувача
func GetUserTickets(pool *pgxpool.Pool, userID int) ([]map[string]interface{}, error) {
	query := `
		SELECT t.id, t.subject, t.status, t.created_at,
			COALESCE(
				(SELECT json_agg(
					json_build_object(
						'id', m.id, 'sender_id', m.sender_id, 
						'message', m.message, 'created_at', m.created_at
					) ORDER BY m.created_at ASC
				) FROM support_messages m WHERE m.ticket_id = t.id),
				'[]'::json
			) AS messages
		FROM support_tickets t WHERE t.user_id = $1 ORDER BY t.created_at DESC
	`
	rows, err := pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту тікетів: %w", err)
	}
	defer rows.Close()

	tickets := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var subject, status string
		var createdAt time.Time
		var messagesJSON []byte
		if err := rows.Scan(&id, &subject, &status, &createdAt, &messagesJSON); err != nil {
			return nil, err
		}

		var messages []map[string]interface{}
		if err := json.Unmarshal(messagesJSON, &messages); err != nil {
			return nil, fmt.Errorf("помилка парсингу повідомлень: %w", err)
		}

		tickets = append(tickets, map[string]interface{}{
			"id":         id,
			"subject":    subject,
			"status":     status,
			"created_at": createdAt,
			"messages":   messages,
		})
	}
	return tickets, nil
}

// GetAllTicketsAdmin повертає всі тікети для панелі адміністратора
func GetAllTicketsAdmin(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `
		SELECT t.id, t.user_id, u.email as user_email, t.subject, t.status, t.created_at,
			COALESCE(
				(SELECT json_agg(
					json_build_object(
						'id', m.id, 'sender_id', m.sender_id, 
						'message', m.message, 'created_at', m.created_at
					) ORDER BY m.created_at ASC
				) FROM support_messages m WHERE m.ticket_id = t.id),
				'[]'::json
			) AS messages
		FROM support_tickets t
		JOIN users u ON t.user_id = u.id
		ORDER BY CASE WHEN t.status = 'open' THEN 1 WHEN t.status = 'in_progress' THEN 2 ELSE 3 END, t.created_at DESC
	`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту всіх тікетів: %w", err)
	}
	defer rows.Close()

	tickets := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, userID int
		var userEmail, subject, status string
		var createdAt time.Time
		var messagesJSON []byte
		if err := rows.Scan(&id, &userID, &userEmail, &subject, &status, &createdAt, &messagesJSON); err != nil {
			return nil, err
		}

		var messages []map[string]interface{}
		if err := json.Unmarshal(messagesJSON, &messages); err != nil {
			return nil, fmt.Errorf("помилка парсингу повідомлень: %w", err)
		}

		tickets = append(tickets, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"user_email": userEmail,
			"subject":    subject,
			"status":     status,
			"created_at": createdAt,
			"messages":   messages,
		})
	}
	return tickets, nil
}

// UpdateTicketStatus змінює статус тікета
func UpdateTicketStatus(pool *pgxpool.Pool, ticketID int, status string) error {
	cmdTag, err := pool.Exec(context.Background(), "UPDATE support_tickets SET status = $1 WHERE id = $2", status, ticketID)
	if err != nil {
		return fmt.Errorf("помилка оновлення статусу тікета: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("тікет не знайдено")
	}
	return nil
}

// AddMessageToTicket додає нове повідомлення до існуючого тікета
func AddMessageToTicket(pool *pgxpool.Pool, ticketID, senderID int, message string) error {
	query := `INSERT INTO support_messages (ticket_id, sender_id, message) VALUES ($1, $2, $3)`
	if _, err := pool.Exec(context.Background(), query, ticketID, senderID, message); err != nil {
		return fmt.Errorf("помилка додавання повідомлення: %w", err)
	}
	return nil
}

// Заглушки для розширення (деталі повідомлень тікета) можна додати пізніше, якщо потрібно розгорнути повноцінний чат.

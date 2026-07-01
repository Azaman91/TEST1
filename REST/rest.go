package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func InitDB() error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		return fmt.Errorf("БД не запустилась %w", err)
	}
	db = pool
	return nil
}

type Subscribe struct {
	ID         string `db:"id"`
	Name       string `json:"service_name" db:"servic_name"`
	Price      int    `json:"price" db:"price"`
	User_id    string `json:"user_id" db:"user_id"`
	Start_date string `json:"start_date" db:"start_date"`
	End_date   string `db:"end_date"`
}

// заполнение данных с запроса подписки
func SubHandler(w http.ResponseWriter, r *http.Request) {
	read, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error Read Body", 400)
		return
	}
	defer r.Body.Close()

	var sub Subscribe
	err = json.Unmarshal(read, &sub)
	if err != nil {
		http.Error(w, "Error Unmarshal Json", 400)
		return
	}

	startDate, err := time.Parse("01-2006", sub.Start_date)
	if err != nil {
		http.Error(w, "Error parse date, expected MM-YYYY", 400)
		return
	}
	sub.End_date = startDate.AddDate(0, 1, 0).Format("01-2006")

	ctx := context.Background()

	var id string
	err = db.QueryRow(ctx, `
        INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `, sub.Name, sub.Price, sub.User_id, sub.Start_date, sub.End_date).Scan(&id)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create subscription: %v", err), 500)
		return
	}

	sub.ID = id

	data, err := json.Marshal(&sub)
	if err != nil {
		http.Error(w, "Error json Marshal", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

// чтение всех данных
func ListSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	rows, err := db.Query(ctx, `
        SELECT id, service_name, price, user_id, start_date, end_date
        FROM subscriptions
    `)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	defer rows.Close()

	var subscriptions []Subscribe

	for rows.Next() {
		var sub Subscribe
		err := rows.Scan(
			&sub.ID, &sub.Name, &sub.Price, &sub.User_id,
			&sub.Start_date, &sub.End_date,
		)
		if err != nil {
			continue
		}
		subscriptions = append(subscriptions, sub)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(subscriptions)
	}
}

// чтение данных
func GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var sub Subscribe
	ctx := context.Background()

	err := db.QueryRow(ctx, `
        SELECT id, service_name, price, user_id, start_date, end_date
        FROM subscriptions 
		WHERE id = $1
    `, id).Scan(&sub.ID, &sub.Name, &sub.Price, &sub.User_id, &sub.Start_date, &sub.End_date)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	data, err := json.Marshal(&sub)
	if err != nil {
		http.Error(w, "Error json Marshal", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

// обновление данных
func UpdateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	read, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error Read Body", 400)
		return
	}
	var sub Subscribe
	err = json.Unmarshal(read, &sub)

	if err != nil {
		http.Error(w, "Error Unmarshal Json", 400)
		return
	}

	ctx := context.Background()

	_, err = db.Exec(ctx, `
        UPDATE subscriptions
        SET service_name = $1,price = $2, start_date = $3,end_date = $4
        WHERE id = $5
    `, sub.Name, sub.Price, sub.Start_date, sub.End_date, id)

	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	w.WriteHeader(200)
}

// удаление данных
func DeleteSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := context.Background()
	_, err := db.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// подсчет суммы
func TotalCostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	read, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error Read Body", 400)
		return
	}
	defer r.Body.Close()

	var req struct {
		Start_date string `json:"start_date"`
		End_date   string `json:"end_date"`
		User_id    string `json:"user_id,omitempty"`
		Name       string `json:"service_name,omitempty"`
	}

	if err := json.Unmarshal(read, &req); err != nil {
		http.Error(w, "Error Unmarshal Json", 400)
		return
	}

	if req.Start_date == "" {
		http.Error(w, "start_date is required", 400)
		return
	}

	if req.End_date == "" {
		req.End_date = req.Start_date
	}

	ctx := context.Background()

	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE start_date = $1 
		  AND (end_date = $2 OR end_date IS NULL)
	`
	args := []interface{}{req.Start_date, req.End_date}
	argIndex := 3

	if req.User_id != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, req.User_id)
		argIndex++
	}

	if req.Name != "" {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argIndex)
		args = append(args, "%"+req.Name+"%")
		argIndex++
	}

	var total int
	err = db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_cost": total,
		"period":     req.Start_date + " - " + req.End_date,
		"filters": map[string]string{
			"user_id":      req.User_id,
			"service_name": req.Name,
		},
	})
}

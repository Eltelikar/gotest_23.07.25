package postgre

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type RequestFields struct {
	ServiceName string     `json:"service_name"`
	Price       uint16     `json:"price"`
	UserId      string     `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

type Storage struct {
	db *sql.DB
}

func New(storageLink string) (*Storage, error) {
	const op = "internal.postgre.New"

	db, err := sql.Open("postgre", storageLink)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db.SetMaxOpenConns(50)
	db.SetConnMaxIdleTime(20)
	db.SetConnMaxLifetime(30 * time.Minute)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS subscriptions(
		service_name TEXT NOT NULL,
		price BIGINT CHECK (price > 0),
		user_id UUID NOT NULL,
		start_date DATE NOT NULL,
		end_date DATE,
		PRIMARY KEY (service_name, user_id));
		`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create table: %w", op, err)
	}

	_, err = db.Exec(`
	CREATE INDEX IF NOT EXISTS idx_subscription ON subscriptions(service_name)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create index: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// Create создает новую запись о подписке в таблице.
func (s *Storage) Create(rb RequestFields) (string, error) {
	const op = "internal.postgre.Create"
	slog.Info("Start create tx", slog.String("op", op))

	tx, err := s.db.Begin()
	if err != nil {
		return "", fmt.Errorf("%s: failed to begin tx: %w", op, err)
	}
	defer rollback(tx, op)

	res, err := tx.Exec(`
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES($1, $2, $3, $4, $5)
		ON CONFLICT (service_name, user_id) DO NOTHING
	`, rb.ServiceName, rb.Price, rb.UserId, rb.StartDate, rb.EndDate)
	if err != nil {
		return "", fmt.Errorf("%s: failed to insert into table: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("%s: failed to read sql result: %w", op, err)
	}

	if rowsAffected == 0 {
		slog.Info("Subsctibtion already exists", slog.String("service_name", rb.ServiceName), slog.String("user_id", rb.UserId))
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	slog.Info("Create done successfully", slog.String("op", op))
	return rb.UserId, nil
}

// Read возвращает информацию о подписке по имени сервиса и ID пользователя.
func (s *Storage) Read(service_name, user_id string) (*RequestFields, error) {
	const op = "internal.postgre.Read"
	slog.Info("Start read tx", slog.String("op", op))

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to begin tx: %w", op, err)
	}
	defer rollback(tx, op)

	var rb RequestFields

	err = tx.QueryRow(`
		SELECT service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE service_name = $1 AND user_id = $2
	`, service_name, user_id).Scan(&rb.ServiceName, &rb.Price, &rb.UserId, &rb.StartDate, &rb.EndDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: no subscription found for service %s and user %s", op, service_name, user_id)
		}
		return nil, fmt.Errorf("%s: failed to query row: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	slog.Info("Read done successfully", slog.String("op", op))
	return &rb, nil
}

// Update обновляет информацию о подписке в таблице.
func (s *Storage) Update(rb RequestFields) error {
	const op = "internal.postgre.Update"
	slog.Info("Start update tx", slog.String("op", op))

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: failed to begin tx: %w", op, err)
	}
	defer rollback(tx, op)

	res, err := tx.Exec(`
		UPDATE subscriptions
		SET price = $1, start_date = $2, end_date = $3
		WHERE service_name = $4 AND user_id = $5
	`, rb.Price, rb.StartDate, rb.EndDate, rb.ServiceName, rb.UserId)
	if err != nil {
		return fmt.Errorf("%s: failed to update table: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to read sql result: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no subscription found for service %s and user %s", op, rb.ServiceName, rb.UserId)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	slog.Info("Update done successfully", slog.String("op", op))
	return nil
}

// Delete удаляет запись о подписке из таблицы.
func (s *Storage) Delete(service_name, user_id string) error {
	const op = "internal.postgre.Delete"
	slog.Info("Start delete tx", slog.String("op", op))

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: failed to begin tx: %w", op, err)
	}
	defer rollback(tx, op)

	res, err := tx.Exec(`
		DELETE FROM subscriptions
		WHERE service_name = $1 AND user_id = $2
	`, service_name, user_id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete from table: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to read sql result: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no subscription found for service %s and user %s", op, service_name, user_id)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	slog.Info("Delete done successfully", slog.String("op", op))
	return nil
}

// List возвращает список всех подписок в таблице.
func (s *Storage) List() ([]RequestFields, error) {
	const op = "internal.postgre.List"
	slog.Info("Start list tx", slog.String("op", op))

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to begin tx: %w", op, err)
	}
	defer rollback(tx, op)

	rows, err := tx.Query(`
		SELECT service_name, price, user_id, start_date, end_date
		FROM subscriptions
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query rows: %w", op, err)
	}
	defer rows.Close()

	var subscriptions []RequestFields

	for rows.Next() {
		var rb RequestFields
		if err := rows.Scan(&rb.ServiceName, &rb.Price, &rb.UserId, &rb.StartDate, &rb.EndDate); err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}
		subscriptions = append(subscriptions, rb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows scan error: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	slog.Info("List done successfully", slog.String("op", op))
	return subscriptions, nil
}

// RangePrice возвращает общую стоимость подписок за указанный диапазон дат и по указанным имени сервиса и id пользователя.
func (s *Storage) RangePrice(start_date time.Time, end_date time.Time, service_name string, user_id string) (uint64, error) {
	const op = "internal.postgre.RangePrice"
	slog.Info("Start range price tx", slog.String("op", op))

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to begin tx: %w", op, err)
	}
	defer rollback(tx, op)

	var totalPrice uint64

	err = tx.QueryRow(`
		SELECT SUM(price)
		FROM subscriptions
		WHERE start_date >= $1 AND end_date <= $2 AND service_name = $3 AND user_id = $4
	`, start_date, end_date, service_name, user_id).Scan(&totalPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("%s: no subscriptions found in the specified date range", op)
		}
		return 0, fmt.Errorf("%s: failed to query total price: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	slog.Info("Range price done successfully", slog.String("op", op))
	return totalPrice, nil
}

func (s *Storage) Close() error {
	const op = "internal.postgre.Close"
	slog.Info("Start close db connection", slog.String("op", op))

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s: failed to close db connection: %w", op, err)
	}

	slog.Info("Close db connection done successfully", slog.String("op", op))
	return nil
}

func rollback(tx *sql.Tx, op string) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		slog.Error("Failed to rollback tx", slog.String("op", op), slog.Any("error", err))
	}
}

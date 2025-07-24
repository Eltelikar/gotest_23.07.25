package postgre

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type RequestFields struct {
	service_name string
	price        uint16
	user_id      string
	start_date   time.Time
	end_date     *time.Time
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
	`, rb.service_name, rb.price, rb.user_id, rb.start_date, rb.end_date)
	if err != nil {
		return "", fmt.Errorf("%s: failed to insert into table: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("%s: failed to read sql result: %w", op, err)
	}

	if rowsAffected == 0 {
		slog.Info("Subsctibtion already exists", slog.String("service_name", rb.service_name), slog.String("user_id", rb.user_id))
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	slog.Info("Create done successfully", slog.String("op", op))
	return rb.user_id, nil
}

//TODO: read subscribe - получить одну подписку
//TODO: update subscribe
//TODO: delete subscribe
//TODO: list subscribes - получить все подписки
//TODO: range price - получить общую стоимость подписок в указанном диапазоне дат

func rollback(tx *sql.Tx, op string) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		slog.Error("Failed to rollback tx", slog.String("op", op), slog.Any("error", err))
	}
}

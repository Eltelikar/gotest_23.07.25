package postgre

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

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

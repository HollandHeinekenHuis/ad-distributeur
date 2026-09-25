package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUnavailable is returned when requested dates overlap an existing booking.
var ErrUnavailable = errors.New("dates unavailable")

// Booking is a stored reservation.
type Booking struct {
	ID        int64
	Name      string
	Email     string
	Phone     string
	Guests    int
	CheckIn   time.Time
	CheckOut  time.Time
	Message   string
	CreatedAt time.Time
}

// Nights returns the number of nights of the stay.
func (b Booking) Nights() int {
	return int(b.CheckOut.Sub(b.CheckIn).Hours() / 24)
}

// DateRange is a booked period (check-out day is free for a new check-in).
type DateRange struct {
	CheckIn  string `json:"checkIn"`
	CheckOut string `json:"checkOut"`
}

type Store struct{ pool *pgxpool.Pool }

func newStore(ctx context.Context, url string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 5
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

const schema = `
CREATE TABLE IF NOT EXISTS bookings (
  id         BIGSERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  email      TEXT NOT NULL,
  phone      TEXT NOT NULL DEFAULT '',
  guests     INT  NOT NULL,
  check_in   DATE NOT NULL,
  check_out  DATE NOT NULL,
  message    TEXT NOT NULL DEFAULT '',
  status     TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (check_out > check_in)
);
CREATE INDEX IF NOT EXISTS bookings_dates_idx ON bookings (check_in, check_out);
`

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, schema)
	return err
}

func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// CreateBooking inserts b unless its dates overlap a non-cancelled booking.
func (s *Store) CreateBooking(ctx context.Context, b *Booking) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Serialise bookings so two concurrent requests cannot take the same dates.
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(7421)"); err != nil {
		return err
	}
	var n int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM bookings
		WHERE status <> 'cancelled' AND check_in < $2 AND check_out > $1`,
		b.CheckIn, b.CheckOut).Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrUnavailable
	}
	err = tx.QueryRow(ctx, `INSERT INTO bookings
		(name, email, phone, guests, check_in, check_out, message)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`,
		b.Name, b.Email, b.Phone, b.Guests, b.CheckIn, b.CheckOut, b.Message,
	).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// BookedRanges lists upcoming booked periods.
func (s *Store) BookedRanges(ctx context.Context) ([]DateRange, error) {
	rows, err := s.pool.Query(ctx, `SELECT check_in, check_out FROM bookings
		WHERE status <> 'cancelled' AND check_out >= CURRENT_DATE ORDER BY check_in`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DateRange{}
	for rows.Next() {
		var in, outD time.Time
		if err := rows.Scan(&in, &outD); err != nil {
			return nil, err
		}
		out = append(out, DateRange{in.Format(dateLayout), outD.Format(dateLayout)})
	}
	return out, rows.Err()
}

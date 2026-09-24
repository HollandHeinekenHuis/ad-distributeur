package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type server struct {
	cfg    Config
	store  *Store
	mailer Mailer
	ready  atomic.Bool
}

func (s *server) routes(static fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", s.handleReady)
	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("GET /api/availability", s.handleAvailability)
	mux.HandleFunc("POST /api/bookings", s.handleCreateBooking)
	mux.Handle("GET /", http.FileServerFS(static))
	return mux
}

// migrateLoop retries the schema migration until the database is reachable.
func (s *server) migrateLoop(ctx context.Context) {
	for {
		mctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := s.store.Migrate(mctx)
		cancel()
		if err == nil {
			s.ready.Store(true)
			log.Println("database ready")
			return
		}
		log.Printf("database not ready yet: %v", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (s *server) dbReady() bool { return s.store != nil && s.ready.Load() }

func (s *server) handleReady(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		w.Write([]byte("ok (no database)"))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if !s.ready.Load() || s.store.Ping(ctx) != nil {
		http.Error(w, "database not ready", http.StatusServiceUnavailable)
		return
	}
	w.Write([]byte("ok"))
}

func (s *server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"villaName":     s.cfg.VillaName,
		"pricePerNight": s.cfg.PricePerNight,
		"currency":      s.cfg.Currency,
		"maxGuests":     s.cfg.MaxGuests,
		"minNights":     s.cfg.MinNights,
	})
}

func (s *server) handleAvailability(w http.ResponseWriter, r *http.Request) {
	if !s.dbReady() {
		writeJSON(w, http.StatusOK, map[string]any{"booked": []DateRange{}})
		return
	}
	ranges, err := s.store.BookedRanges(r.Context())
	if err != nil {
		log.Printf("availability: %v", err)
		writeError(w, http.StatusInternalServerError, "could not load availability")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"booked": ranges})
}

func (s *server) handleCreateBooking(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	b, err := req.Validate(time.Now(), s.cfg)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !s.dbReady() {
		writeError(w, http.StatusServiceUnavailable, "booking is temporarily unavailable, please try again shortly")
		return
	}
	if err := s.store.CreateBooking(r.Context(), &b); err != nil {
		if errors.Is(err, ErrUnavailable) {
			writeError(w, http.StatusConflict, "sorry, those dates are already booked")
			return
		}
		log.Printf("create booking: %v", err)
		writeError(w, http.StatusInternalServerError, "could not save booking")
		return
	}
	go s.sendBookingEmails(b)
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":     b.ID,
		"nights": b.Nights(),
		"total":  b.Nights() * s.cfg.PricePerNight,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

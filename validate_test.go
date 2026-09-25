package main

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	cfg := Config{MaxGuests: 8, MinNights: 2}
	now := time.Date(2030, 5, 1, 12, 0, 0, 0, time.UTC)
	ok := BookingRequest{Name: "Anna", Email: "anna@example.com", Guests: 4,
		CheckIn: "2030-06-01", CheckOut: "2030-06-08"}

	b, err := ok.Validate(now, cfg)
	if err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	if b.Nights() != 7 {
		t.Fatalf("nights = %d, want 7", b.Nights())
	}

	cases := map[string]func(r *BookingRequest){
		"no name":     func(r *BookingRequest) { r.Name = " " },
		"bad email":   func(r *BookingRequest) { r.Email = "nope" },
		"too many":    func(r *BookingRequest) { r.Guests = 9 },
		"past":        func(r *BookingRequest) { r.CheckIn = "2030-04-01" },
		"too short":   func(r *BookingRequest) { r.CheckOut = "2030-06-02" },
		"reversed":    func(r *BookingRequest) { r.CheckOut = "2030-05-20" },
		"bad date":    func(r *BookingRequest) { r.CheckIn = "01/06/2030" },
	}
	for name, mutate := range cases {
		r := ok
		mutate(&r)
		if _, err := r.Validate(now, cfg); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestSanitizeHeader(t *testing.T) {
	if got := sanitizeHeader("a\r\nBcc: x@y.z"); got != "aBcc: x@y.z" {
		t.Fatalf("got %q", got)
	}
}

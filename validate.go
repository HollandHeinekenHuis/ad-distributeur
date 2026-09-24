package main

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// BookingRequest is the JSON body of POST /api/bookings.
type BookingRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Guests   int    `json:"guests"`
	CheckIn  string `json:"checkIn"`
	CheckOut string `json:"checkOut"`
	Message  string `json:"message"`
}

// Validate checks the request and converts it into a Booking.
func (r BookingRequest) Validate(now time.Time, cfg Config) (Booking, error) {
	name := strings.TrimSpace(r.Name)
	if name == "" || len(name) > 200 {
		return Booking{}, errors.New("please enter your name")
	}
	addr, err := mail.ParseAddress(strings.TrimSpace(r.Email))
	if err != nil || len(addr.Address) > 254 {
		return Booking{}, errors.New("please enter a valid email address")
	}
	phone := strings.TrimSpace(r.Phone)
	if len(phone) > 40 {
		return Booking{}, errors.New("phone number is too long")
	}
	msg := strings.TrimSpace(r.Message)
	if len(msg) > 2000 {
		return Booking{}, errors.New("message is too long (max 2000 characters)")
	}
	if r.Guests < 1 || r.Guests > cfg.MaxGuests {
		return Booking{}, errors.New("invalid number of guests")
	}
	in, err1 := time.Parse(dateLayout, r.CheckIn)
	out, err2 := time.Parse(dateLayout, r.CheckOut)
	if err1 != nil || err2 != nil {
		return Booking{}, errors.New("please choose check-in and check-out dates")
	}
	today, _ := time.Parse(dateLayout, now.UTC().Format(dateLayout))
	if in.Before(today) {
		return Booking{}, errors.New("check-in cannot be in the past")
	}
	b := Booking{Name: name, Email: addr.Address, Phone: phone, Guests: r.Guests,
		CheckIn: in, CheckOut: out, Message: msg}
	if b.Nights() < cfg.MinNights {
		return Booking{}, errors.New("stay is shorter than the minimum number of nights")
	}
	if b.Nights() > 90 {
		return Booking{}, errors.New("stays longer than 90 nights must be arranged by email")
	}
	return b, nil
}

package main

import (
	"os"
	"strconv"
)

// Config holds all runtime settings, read from environment variables.
type Config struct {
	Port          string
	DatabaseURL   string
	VillaName     string
	PricePerNight int
	Currency      string
	MaxGuests     int
	MinNights     int

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	MailFrom     string
	OwnerEmail   string
}

func loadConfig() Config {
	return Config{
		Port:          envOr("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		VillaName:     envOr("VILLA_NAME", "Villa Les Oliviers"),
		PricePerNight: envInt("PRICE_PER_NIGHT", 350),
		Currency:      envOr("CURRENCY", "EUR"),
		MaxGuests:     envInt("MAX_GUESTS", 8),
		MinNights:     envInt("MIN_NIGHTS", 2),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      envOr("SMTP_PORT", "587"),
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		MailFrom:      envOr("MAIL_FROM", "bookings@example.com"),
		OwnerEmail:    os.Getenv("OWNER_EMAIL"),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
		return v
	}
	return def
}

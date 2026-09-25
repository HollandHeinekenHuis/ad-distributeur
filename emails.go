package main

import (
	"fmt"
	"log"
)

const humanDate = "Monday 2 January 2006"

// sendBookingEmails notifies the guest and (if configured) the owner.
func (s *server) sendBookingEmails(b Booking) {
	total := b.Nights() * s.cfg.PricePerNight
	summary := fmt.Sprintf(
		"Booking reference: #%d\nCheck-in:  %s\nCheck-out: %s\nNights:    %d\nGuests:    %d\nEstimated total: %d %s\n",
		b.ID, b.CheckIn.Format(humanDate), b.CheckOut.Format(humanDate),
		b.Nights(), b.Guests, total, s.cfg.Currency)

	guestBody := fmt.Sprintf("Dear %s,\n\nThank you for your booking request at %s.\n"+
		"We will confirm your stay shortly.\n\n%s\nÀ bientôt!\n%s\n",
		b.Name, s.cfg.VillaName, summary, s.cfg.VillaName)
	if err := s.mailer.Send(b.Email, "Your booking request at "+s.cfg.VillaName, guestBody); err != nil {
		log.Printf("guest email for booking #%d failed: %v", b.ID, err)
	}

	if s.cfg.OwnerEmail == "" {
		return
	}
	ownerBody := fmt.Sprintf("New booking request\n\n%s\nName:  %s\nEmail: %s\nPhone: %s\n\nMessage:\n%s\n",
		summary, b.Name, b.Email, b.Phone, b.Message)
	subject := fmt.Sprintf("New booking #%d: %s", b.ID, b.CheckIn.Format(dateLayout))
	if err := s.mailer.Send(s.cfg.OwnerEmail, subject, ownerBody); err != nil {
		log.Printf("owner email for booking #%d failed: %v", b.ID, err)
	}
}

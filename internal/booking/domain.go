package booking

import (
	"errors"
	"time"
)

var (
	ErrSeatAlreadyBooked = errors.New("seat is already taken")
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionExpired    = errors.New("session has expired")
)

// Booking represents a confirmed seat reservation.
type Booking struct {
	ID        string    `json:"id"`
	MovieID   string    `json:"movie_id"`
	SeatID    string    `json:"seat_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"` // "held" or "confirmed"
	ExpiresAt time.Time `json:"expires_at"`
}

type BookingStore interface {
	Book(b Booking) (Booking, error)
	ListBookings(movieID string) []Booking
	Confirm(sessionID string) error
	Release(sessionID string) error
}

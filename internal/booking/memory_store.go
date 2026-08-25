package booking

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu       sync.RWMutex
	bookings map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: map[string]Booking{},
	}
}

func (s *MemoryStore) Book(b Booking) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if seat is already booked/held
	if existing, exists := s.bookings[b.SeatID]; exists {
		// If the booking is confirmed, or if it is held and hasn't expired yet:
		if existing.Status == "confirmed" || (existing.Status == "held" && time.Now().Before(existing.ExpiresAt)) {
			return Booking{}, ErrSeatAlreadyBooked
		}
	}

	// Generate new session metadata
	b.ID = uuid.New().String()
	b.ExpiresAt = time.Now().Add(2 * time.Minute) // Hold seat for 2 minutes

	s.bookings[b.SeatID] = b

	return b, nil
}

func (s *MemoryStore) ListBookings(movieID string) []Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Booking
	for _, b := range s.bookings {
		if b.MovieID == movieID {
			// Skip returning holds that have expired
			if b.Status == "held" && time.Now().After(b.ExpiresAt) {
				continue
			}
			result = append(result, b)
		}
	}
	return result
}

func (s *MemoryStore) Confirm(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the booking with this sessionID
	for seatID, b := range s.bookings {
		if b.ID == sessionID {
			// If it expired, delete it and return an error
			if b.Status == "held" && time.Now().After(b.ExpiresAt) {
				delete(s.bookings, seatID)
				return ErrSessionExpired
			}

			// Confirm the booking
			b.Status = "confirmed"
			b.ExpiresAt = time.Time{} // Clear expiration time
			s.bookings[seatID] = b
			return nil
		}
	}

	return ErrSessionNotFound
}

func (s *MemoryStore) Release(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and delete the held booking
	for seatID, b := range s.bookings {
		if b.ID == sessionID {
			delete(s.bookings, seatID)
			return nil
		}
	}

	return ErrSessionNotFound
}

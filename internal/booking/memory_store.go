package booking

import "sync"

type MemoryStore struct {
	mu       sync.RWMutex
	bookings map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: map[string]Booking{},
	}
}

func (s *MemoryStore) Book(b Booking) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.bookings[b.SeatID]; exists {
		return ErrSeatAlreadyBooked
	}
	s.bookings[b.SeatID] = b

	return nil
}

func (s *MemoryStore) ListBookings(movieID string) []Booking {

	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Booking

	for _, b := range s.bookings {
		if b.MovieID == movieID {
			result = append(result, b)
		}
	}
	return result
}

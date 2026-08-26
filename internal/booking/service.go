package booking

type Service struct {
	store BookingStore
}

func NewService(store BookingStore) *Service {
	return &Service{store}
}

// Book delegates booking creation to the store layer and returns the generated booking details
func (s *Service) Book(b Booking) (Booking, error) {
	return s.store.Book(b)
}

// Confirm delegates the session confirmation to the store layer
func (s *Service) Confirm(sessionID string) error {
	return s.store.Confirm(sessionID)
}

// Release delegates the session release to the store layer
func (s *Service) Release(sessionID string) error {
	return s.store.Release(sessionID)
}

package booking

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/abhishekaringale/concurrent-cinema-booking/internal/utils"
)

// sseBroker manages active client channels for a specific movie screening
type sseBroker struct {
	mu        sync.RWMutex
	listeners map[chan []SeatStatus]bool
}

func newBroker() *sseBroker {
	return &sseBroker{
		listeners: make(map[chan []SeatStatus]bool),
	}
}

func (b *sseBroker) add(ch chan []SeatStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners[ch] = true
}

func (b *sseBroker) remove(ch chan []SeatStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.listeners, ch)
	close(ch)
}

func (b *sseBroker) broadcast(seats []SeatStatus) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.listeners {
		select {
		case ch <- seats:
		default: // avoid blocking if a client has slow network/buffer is full
		}
	}
}

type handler struct {
	svc       *Service
	brokers   map[string]*sseBroker
	brokersMu sync.Mutex
}

func NewHandler(svc *Service) *handler {
	return &handler{
		svc:     svc,
		brokers: make(map[string]*sseBroker),
	}
}

func (h *handler) getBroker(movieID string) *sseBroker {
	h.brokersMu.Lock()
	defer h.brokersMu.Unlock()
	b, ok := h.brokers[movieID]
	if !ok {
		b = newBroker()
		h.brokers[movieID] = b
	}
	return b
}

type SeatStatus struct {
	SeatID    string `json:"seat_id"`
	Booked    bool   `json:"booked"`
	Confirmed bool   `json:"confirmed"`
	UserID    string `json:"user_id"`
}

// broadcastUpdate retrieves the current seat statuses and pushes them to the broker
func (h *handler) broadcastUpdate(movieID string) {
	bookings := h.svc.store.ListBookings(movieID)
	seatStatusUses := []SeatStatus{}
	for _, b := range bookings {
		seatStatusUses = append(seatStatusUses, SeatStatus{
			SeatID:    b.SeatID,
			Booked:    true,
			Confirmed: b.Status == "confirmed",
			UserID:    b.UserID,
		})
	}
	h.getBroker(movieID).broadcast(seatStatusUses)
}

// ListSeats responds with the current status of all seats for a movie
func (h *handler) ListSeats(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")

	bookings := h.svc.store.ListBookings(movieID)
	seatStatusUses := []SeatStatus{}

	for _, b := range bookings {
		seatStatusUses = append(seatStatusUses, SeatStatus{
			SeatID:    b.SeatID,
			Booked:    true,
			Confirmed: b.Status == "confirmed",
			UserID:    b.UserID,
		})
	}

	utils.WriteJSON(w, http.StatusOK, seatStatusUses)
}

type holdRequest struct {
	UserID string `json:"user_id"`
}

// HoldSeat handles the request to temporarily hold a seat and broadcasts the update
func (h *handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	seatID := r.PathValue("seatID")

	var req holdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	b := Booking{
		MovieID: movieID,
		SeatID:  seatID,
		UserID:  req.UserID,
		Status:  "held",
	}

	session, err := h.svc.Book(b)
	if err != nil {
		if err == ErrSeatAlreadyBooked {
			utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Broadcast the new hold layout dynamically to other users
	h.broadcastUpdate(movieID)

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id": session.ID,
		"movie_id":   session.MovieID,
		"seat_id":    session.SeatID,
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
	})
}

// Confirm handles verifying a session ID, locking it in permanently, and broadcasting the update
func (h *handler) Confirm(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	movieID := r.URL.Query().Get("movie_id") // Read movie_id to target target broker room

	if err := h.svc.Confirm(sessionID); err != nil {
		if err == ErrSessionExpired {
			utils.WriteJSON(w, http.StatusGone, map[string]string{"error": err.Error()})
			return
		}
		if err == ErrSessionNotFound {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Broadcast the permanent booking to other screens
	if movieID != "" {
		h.broadcastUpdate(movieID)
	}

	w.WriteHeader(http.StatusNoContent)
}

// Release handles deleting a held session to make the seat free, then broadcasts the update
func (h *handler) Release(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	movieID := r.URL.Query().Get("movie_id") // Read movie_id to target correct broker room

	if err := h.svc.Release(sessionID); err != nil {
		if err == ErrSessionNotFound {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Broadcast that this seat is now free
	if movieID != "" {
		h.broadcastUpdate(movieID)
	}

	w.WriteHeader(http.StatusNoContent)
}

// StreamSeats creates a persistent SSE connection to stream seat layout updates in real-time
func (h *handler) StreamSeats(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	if movieID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing movie ID"})
		return
	}

	// Set required SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Create listener channel and register to the broker
	ch := make(chan []SeatStatus, 10)
	broker := h.getBroker(movieID)
	broker.add(ch)
	defer broker.remove(ch)

	// Send initial grid states immediately on connection
	bookings := h.svc.store.ListBookings(movieID)
	seatStatusUses := []SeatStatus{}
	for _, b := range bookings {
		seatStatusUses = append(seatStatusUses, SeatStatus{
			SeatID:    b.SeatID,
			Booked:    true,
			Confirmed: b.Status == "confirmed",
			UserID:    b.UserID,
		})
	}
	initData, _ := json.Marshal(seatStatusUses)
	fmt.Fprintf(w, "data: %s\n\n", initData)
	flusher.Flush()

	// Wait for client disconnection or broadcasted updates
	for {
		select {
		case <-r.Context().Done():
			// Clean exit when client closes tab/navigates away
			return
		case seats := <-ch:
			data, err := json.Marshal(seats)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

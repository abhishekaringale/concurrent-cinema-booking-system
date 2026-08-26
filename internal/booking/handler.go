package booking

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/abhishekaringale/concurrent-cinema-booking/internal/utils"
)

type handler struct {
	svc *Service
}

func NewHandler(svc *Service) *handler {
	return &handler{svc: svc}
}

type SeatStatus struct {
	SeatID    string `json:"seat_id"`
	Booked    bool   `json:"booked"`
	Confirmed bool   `json:"confirmed"`
	UserID    string `json:"user_id"`
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

// HoldSeat handles the request to temporarily hold a seat
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

	// 1. Call service.go instead of the store directly, capturing the new session object
	session, err := h.svc.Book(b)
	if err != nil {
		if err == ErrSeatAlreadyBooked {
			utils.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}

		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// 2. Return the real session details and format ExpiresAt using RFC3339 for the frontend
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id": session.ID,
		"movie_id":   session.MovieID,
		"seat_id":    session.SeatID,
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
	})
}

// Confirm handles verifying a session ID and locking in the seat permanently
func (h *handler) Confirm(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

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

	w.WriteHeader(http.StatusNoContent) // 204 No Content on success
}

// Release handles deleting a held session to make the seat free again
func (h *handler) Release(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	if err := h.svc.Release(sessionID); err != nil {
		if err == ErrSessionNotFound {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content on success
}

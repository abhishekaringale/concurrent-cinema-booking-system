package booking

import (
	"encoding/json"
	"net/http"

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

	if err := h.svc.store.Book(b); err != nil {
		if err == ErrSeatAlreadyBooked {
			utils.WriteJSON(w, http.StatusConflict, map[string]string{
				"error": err.Error()})
			return
		}

		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id": "dummy-session-id",
		"movie_id":   movieID,
		"seat_id":    seatID,
		"expires_at": "2026-08-23T20:00:00Z", //dummy
	})
}

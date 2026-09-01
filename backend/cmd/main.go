package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/abhishekaringale/concurrent-cinema-booking/internal/booking"
	"github.com/abhishekaringale/concurrent-cinema-booking/internal/utils"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Movie struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Rows        int    `json:"rows"`
	SeatsPerRow int    `json:"seats_per_row"`
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load environment variables from .env file if present
	_ = godotenv.Load()

	// 1. Initialize the Redis client connecting to our local server or Upstash
	redisURL := os.Getenv("REDIS_URL")
	var opt *redis.Options
	var err error
	if redisURL != "" {
		// 1. In Production / Cloud: Parse Upstash URL (supports SSL "rediss://")
		opt, err = redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("failed to parse REDIS_URL: %v", err)
		}
	} else {
		// 2. On Your Laptop: Fallback to local Redis
		opt = &redis.Options{
			Addr: "localhost:6379",
		}
	}
	rdb := redis.NewClient(opt)

	// 2. Ping Redis to verify connection works before starting the server
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	// 3. Initialize our RedisStore
	store := booking.NewRedisStore(rdb)

	svc := booking.NewService(store)
	bookingHandler := booking.NewHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", rootHandler)
	mux.HandleFunc("GET /movies", listMoviesHandler)
	mux.HandleFunc("GET /movies/{movieID}/seats", bookingHandler.ListSeats)
	mux.HandleFunc("GET /movies/{movieID}/seats/stream", bookingHandler.StreamSeats)
	mux.HandleFunc("POST /movies/{movieID}/seats/{seatID}/hold", bookingHandler.HoldSeat)
	mux.HandleFunc("PUT /sessions/{sessionID}/confirm", bookingHandler.Confirm)
	mux.HandleFunc("DELETE /sessions/{sessionID}", bookingHandler.Release)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server is running on port %s...", port)

	if err := http.ListenAndServe(":"+port, enableCORS(mux)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"service": "Concurrent Cinema Booking API",
		"status":  "healthy",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"GET /movies":                                  "List all available movie screenings",
			"GET /movies/{movieID}/seats":                  "Fetch current seat statuses",
			"GET /movies/{movieID}/seats/stream":           "Real-time Server-Sent Events (SSE) stream",
			"POST /movies/{movieID}/seats/{seatID}/hold":   "Temporarily hold a seat for 2 minutes",
			"PUT /sessions/{sessionID}/confirm?movie_id=":  "Permanently confirm a held seat reservation",
			"DELETE /sessions/{sessionID}?movie_id=":       "Release a held seat reservation",
		},
		"github": "https://github.com/abhishekaringale/concurrent-cinema-booking-system",
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	movies := []Movie{
		{ID: "movie-1", Title: "The Matrix", Rows: 5, SeatsPerRow: 8},
		{ID: "movie-2", Title: "Inception", Rows: 6, SeatsPerRow: 10},
		{ID: "movie-3", Title: "Interstellar", Rows: 8, SeatsPerRow: 12},
	}

	utils.WriteJSON(w, http.StatusOK, movies)
}

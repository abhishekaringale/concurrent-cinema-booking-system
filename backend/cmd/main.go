package main

import (
	"context" // NEW: Required for Redis commands
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

	// 3. Initialize our RedisStore (swapped from NewMemoryStore)
	store := booking.NewRedisStore(rdb)

	svc := booking.NewService(store)
	bookingHandler := booking.NewHandler(svc)

	mux := http.NewServeMux()

	mux.Handle("GET /", http.FileServer(http.Dir("static")))
	mux.HandleFunc("GET /movies", listMoviesHandler)

	mux.HandleFunc("GET /movies/{movieID}/seats", bookingHandler.ListSeats)
	mux.HandleFunc("GET /movies/{movieID}/seats/stream", bookingHandler.StreamSeats)
	mux.HandleFunc("POST /movies/{movieID}/seats/{seatID}/hold", bookingHandler.HoldSeat)
	mux.HandleFunc("PUT /sessions/{sessionID}/confirm", bookingHandler.Confirm)
	mux.HandleFunc("DELETE /sessions/{sessionID}", bookingHandler.Release)

	log.Println("server is running on http://localhost:8080...")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	movies := []Movie{
		{ID: "movie-1", Title: "The Matrix", Rows: 5, SeatsPerRow: 8},
		{ID: "movie-2", Title: "Inception", Rows: 6, SeatsPerRow: 10},
		{ID: "movie-3", Title: "Interstellar", Rows: 8, SeatsPerRow: 12},
	}

	utils.WriteJSON(w, http.StatusOK, movies)
}

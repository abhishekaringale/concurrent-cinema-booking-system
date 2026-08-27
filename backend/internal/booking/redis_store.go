package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const defaultHoldTTL = 2 * time.Minute

// RedisStore implements BookingStore backed by Redis.
type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

// seatKey creates a unique key for a seat in a movie
func seatKey(movieID, seatID string) string {
	return fmt.Sprintf("seat:%s:%s", movieID, seatID)
}

// sessionKey creates a reverse lookup key for a session ID
func sessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

func (s *RedisStore) Book(b Booking) (Booking, error) {
	ctx := context.Background()

	// 1. Generate session metadata
	b.ID = uuid.New().String()
	b.ExpiresAt = time.Now().Add(defaultHoldTTL)
	b.Status = "held"

	bookingJSON, err := json.Marshal(b)
	if err != nil {
		return Booking{}, err
	}

	sKey := seatKey(b.MovieID, b.SeatID)
	sessKey := sessionKey(b.ID)

	// 2. Set the seat lock atomically ONLY if it doesn't exist (NX) with a TTL
	success, err := s.rdb.SetNX(ctx, sKey, bookingJSON, defaultHoldTTL).Result()
	if err != nil {
		return Booking{}, err
	}
	if !success {
		// Seat is already held or confirmed
		return Booking{}, ErrSeatAlreadyBooked
	}

	// 3. Store the session reverse lookup pointing to the seat lock
	err = s.rdb.Set(ctx, sessKey, sKey, defaultHoldTTL).Err()
	if err != nil {
		// Clean up the seat lock if the session mapping failed
		s.rdb.Del(ctx, sKey)
		return Booking{}, err
	}

	return b, nil
}

func (s *RedisStore) ListBookings(movieID string) []Booking {
	ctx := context.Background()
	pattern := fmt.Sprintf("seat:%s:*", movieID)

	var bookings []Booking

	// Use SCAN instead of KEYS to avoid blocking Redis performance
	var cursor uint64
	for {
		keys, nextCursor, err := s.rdb.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			break
		}

		for _, key := range keys {
			val, err := s.rdb.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var b Booking
			if err := json.Unmarshal([]byte(val), &b); err == nil {
				bookings = append(bookings, b)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return bookings
}

func (s *RedisStore) Confirm(sessionID string) error {
	ctx := context.Background()
	sessKey := sessionKey(sessionID)

	// 1. Find the seat key from the session mapping
	sKey, err := s.rdb.Get(ctx, sessKey).Result()
	if err == redis.Nil {
		return ErrSessionNotFound
	} else if err != nil {
		return err
	}

	// 2. Retrieve the booking details
	val, err := s.rdb.Get(ctx, sKey).Result()
	if err == redis.Nil {
		return ErrSessionExpired
	} else if err != nil {
		return err
	}

	var b Booking
	if err := json.Unmarshal([]byte(val), &b); err != nil {
		return err
	}

	// 3. Update status to confirmed and clear the expiration time
	b.Status = "confirmed"
	b.ExpiresAt = time.Time{}

	updatedJSON, err := json.Marshal(b)
	if err != nil {
		return err
	}

	// 4. Update seat to be confirmed (No TTL) and delete session key atomically
	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, sKey, updatedJSON, 0) // 0 expiration means persist forever
	pipe.Del(ctx, sessKey)
	_, err = pipe.Exec(ctx)

	return err
}

func (s *RedisStore) Release(sessionID string) error {
	ctx := context.Background()
	sessKey := sessionKey(sessionID)

	// 1. Find the seat key
	sKey, err := s.rdb.Get(ctx, sessKey).Result()
	if err == redis.Nil {
		return ErrSessionNotFound
	} else if err != nil {
		return err
	}

	// 2. Delete both keys atomically
	pipe := s.rdb.TxPipeline()
	pipe.Del(ctx, sKey)
	pipe.Del(ctx, sessKey)
	_, err = pipe.Exec(ctx)

	return err
}

package govalidator

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const DEMO_API_KEY = "demo-api-key-123"
const DEMO_API_CREDITS = 100

// Populates Redis with API keys and credits for testing
//
// Parameters:
//   - apiKey: the API key to seed
//   - credits: the number of credits to associate with the API key
//   - rdb: the Redis client instance
//   - ctx: the context for Redis operations
func seed(apiKey string, credits int, rdb *redis.Client, ctx context.Context) {
	err := rdb.Set(ctx, apiKey, credits, 24*time.Hour).Err()
	if err != nil {
		panic(fmt.Sprintf("Failed to seed Redis: %v", err))
	}
}

// Validates API key and checks credits in Redis
// Returns 200 OK if key is valid and credits are sufficient
// Returns 401 Unauthorized if invalid
// Returns 402 Payment Required if insufficient credits
func main() {

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Container
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	// Initialize root context
	ctx := context.Background()
	seed(DEMO_API_KEY, DEMO_API_CREDITS, rdb, ctx)

	http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve API key from header
		apiKey := r.Header.Get("X-API-Key")

		// Check if API key is missing
		if apiKey == "" {
			http.Error(w, "Missing API key", http.StatusBadRequest)
			return
		}

		// Retrieve the key and value from Redis
		result, err := rdb.Get(ctx, apiKey).Result()
		if err != nil {
			http.Error(w, "Error validating API key", http.StatusInternalServerError)
			return
		}

		// Convert the Redis vlue to an integer
		resCredits, err := strconv.Atoi(result)
		if err != nil {
			http.Error(w, "Stored credits are invalid", http.StatusInternalServerError)
			return
		}

		// Check the API key and credits in Redis
		if resCredits > 0 {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "Insufficient credits", http.StatusPaymentRequired)
			return
		}
	})

	// Start the server
	fmt.Println("Go Validator is running on port 8080...")
	http.ListenAndServe(":8080", nil)

	// When the server is stopped, close the Redis client
	defer rdb.Close()
}

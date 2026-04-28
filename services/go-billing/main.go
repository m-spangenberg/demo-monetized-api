package gobilling

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

// Performs billing operations such as credit ledger updates and usage settlement
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

	http.HandleFunc("/billing", func(w http.ResponseWriter, r *http.Request) {
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

		// Credit Ledger Update Logic (simplified for demo)
		credits, err := strconv.Atoi(result)
		if err != nil {
			http.Error(w, "Invalid credit value", http.StatusInternalServerError)
			return
		}

		// Usage Settlement Logic (simplified for demo)
		// For demonstration, we duduct the value
		if credits > 0 {
	})

	// Start the server
	fmt.Println("Go Billing is running on port 8081...")
	http.ListenAndServe(":8081", nil)

	// When the server is stopped, close the Redis client
	defer rdb.Close()
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/m-spangenberg/demo-monetized-api/pkg/common"
	"github.com/redis/go-redis/v9"
)

type ValidatorHandler struct {
	rdb *redis.Client
}

func (h *ValidatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Retrieve API key from Authorization header
	authHeader := r.Header.Get("Authorization")
	var apiKey string
	if strings.HasPrefix(authHeader, "Bearer ") {
		apiKey = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Fallback to X-API-Key for compatibility if needed, but prioritize Bearer
	if apiKey == "" {
		apiKey = r.Header.Get("X-API-Key")
	}

	// Check if API key is missing
	if apiKey == "" {
		http.Error(w, "Missing API key in Authorization header", http.StatusBadRequest)
		return
	}

	// Retrieve the key and value from Redis using request context
	result, err := h.rdb.Get(r.Context(), apiKey).Result()
	if err != nil {
		if err == redis.Nil {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error validating API key", http.StatusInternalServerError)
		return
	}

	// Convert the Redis value to a float
	resCredits, err := strconv.ParseFloat(result, 64)
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
}

func main() {
	cfg := common.LoadConfig()

	// Initialize Redis client
	rdb, err := common.NewRedisClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	handler := &ValidatorHandler{rdb: rdb}

	http.Handle("/api/v1/validate", handler)

	// Start the server
	fmt.Println("Go Validator is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

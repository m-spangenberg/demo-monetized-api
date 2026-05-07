package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/m-spangenberg/demo-monetized-api/pkg/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type APIHandler struct {
	db  *sql.DB
	rdb *redis.Client
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Standardize Bearer token extraction
	authHeader := r.Header.Get("Authorization")
	var apiKey string
	if strings.HasPrefix(authHeader, "Bearer ") {
		apiKey = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if apiKey == "" {
		apiKey = r.Header.Get("X-API-Key")
	}

	if apiKey == "" {
		http.Error(w, "Missing API key in Authorization header", http.StatusBadRequest)
		return
	}

	// Handle endpoints
	switch {
	case r.URL.Path == "/api/v1/info" && r.Method == http.MethodGet:
		h.handleInfo(w, r, apiKey)
	case r.URL.Path == "/api/v1/health" && r.Method == http.MethodGet:
		h.handleHealth(w, r, apiKey)
	case strings.HasPrefix(r.URL.Path, "/api/v1/work/") && r.Method == http.MethodPost:
		h.handleWork(w, r, apiKey)
	default:
		// Fallback for the old /service endpoint if still needed for backwards compatibility during refactor
		if r.URL.Path == "/api/v1/service" {
			h.handleService(w, r, apiKey)
			return
		}
		http.NotFound(w, r)
	}
}

func (h *APIHandler) handleInfo(w http.ResponseWriter, r *http.Request, apiKey string) {
	balance, _ := h.rdb.Get(r.Context(), apiKey).Float64()
	info := map[string]interface{}{
		"name":        "Demo Monetized API",
		"version":     "0.0.1",
		"description": "A demo API. See more at https://api.domain.tld/docs",
		"balance":     balance,
	}
	w.Header().Set("X-Credits-Usage", "0")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *APIHandler) handleHealth(w http.ResponseWriter, r *http.Request, apiKey string) {
	balance, _ := h.rdb.Get(r.Context(), apiKey).Float64()
	health := map[string]interface{}{
		"status":     "healthy",
		"latency_ms": rand.Intn(100) + 50, // Simulate latency between 50-150ms
		"timestamp":  time.Now().Format(time.RFC3339),
		"uptime":     "72h",
		"balance":    balance,
	}
	w.Header().Set("X-Credits-Usage", "0")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (h *APIHandler) handleWork(w http.ResponseWriter, r *http.Request, apiKey string) {
	level := strings.TrimPrefix(r.URL.Path, "/api/v1/work/")
	cost := 1.0
	switch level {
	case "easy":
		cost = 1.0
	case "medium":
		cost = 5.0
	case "hard":
		cost = 10.0
	default:
		http.Error(w, "Invalid work level", http.StatusBadRequest)
		return
	}

	// In a real Pay-per-use API, the API service processes the request
	// and informs the gateway about the cost via a header.
	w.Header().Set("X-Credits-Usage", fmt.Sprintf("%.2f", cost))
	w.Header().Set("Content-Type", "application/json")

	balance, _ := h.rdb.Get(r.Context(), apiKey).Float64()
	result := map[string]interface{}{
		"credits_used": cost,
		"balance":      balance - cost,
		"status":       "success",
		"timestamp":    time.Now().Format(time.RFC3339),
		"duration_ms":  rand.Intn(200) + 100, // Simulate processing time between 100-300ms
		"result":       "Base64-encoded data string representing " + level + " work",
	}

	json.NewEncoder(w).Encode(result)
}

func (h *APIHandler) handleService(w http.ResponseWriter, r *http.Request, apiKey string) {
	// Re-check credits just in case, though Kong should have handled it
	result, err := h.rdb.Get(r.Context(), apiKey).Result()
	if err != nil {
		http.Error(w, "Error validating API key", http.StatusInternalServerError)
		return
	}

	resCredits, _ := strconv.ParseFloat(result, 64)
	if resCredits > 0 {
		w.Header().Set("X-Credits-Usage", "1.0") // Default cost for generic service
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service operation successful"))
	} else {
		http.Error(w, "Insufficient credits", http.StatusPaymentRequired)
	}
}

func main() {
	cfg := common.LoadConfig()

	// Initialize SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	// Initialize Tables
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"name" TEXT
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal(err)
	}

	// Initialize Redis client using shared logic
	rdb, err := common.NewRedisClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	apiHandler := &APIHandler{db: db, rdb: rdb}

	// Use a prefix to catch all /api/v1/* requests
	http.Handle("/api/v1/", apiHandler)

	// Start the server
	fmt.Println("Go API is running on port 8082...")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}
}

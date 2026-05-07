package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/m-spangenberg/demo-monetized-api/pkg/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type APIHandler struct {
	db  *sql.DB
	rdb *redis.Client
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Retrieve API key from header
	apiKey := r.Header.Get("X-API-Key")

	// Check if API key is missing
	if apiKey == "" {
		http.Error(w, "Missing API key", http.StatusBadRequest)
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

	// Convert the Redis value to an integer
	resCredits, err := strconv.Atoi(result)
	if err != nil {
		http.Error(w, "Stored credits are invalid", http.StatusInternalServerError)
		return
	}

	// Check the API key and credits in Redis
	if resCredits > 0 {
		// Business logic would go here
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service operation successful"))
		return
	} else {
		http.Error(w, "Insufficient credits", http.StatusPaymentRequired)
		return
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

	http.Handle("/api/v1/service", apiHandler)

	// Start the server
	fmt.Println("Go API is running on port 8082...")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/m-spangenberg/demo-monetized-api/pkg/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

const DEMO_API_KEY string = "demo-api-key-123"
const DEMO_API_CREDITS float64 = 100.00

type BillingHandler struct {
	db  *sql.DB
	rdb *redis.Client
}

func (h *BillingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Retrieve API key from Authorization header
	authHeader := r.Header.Get("Authorization")
	var apiKey string
	if strings.HasPrefix(authHeader, "Bearer ") {
		apiKey = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if apiKey == "" {
		apiKey = r.Header.Get("X-API-Key")
	}

	// Check if API key is missing
	if apiKey == "" {
		http.Error(w, "Missing API key", http.StatusBadRequest)
		return
	}

	// Retrieve amount to settle from query parameter
	amountStr := r.URL.Query().Get("amount")
	if amountStr == "" {
		http.Error(w, "Missing amount parameter", http.StatusBadRequest)
		return
	}

	// Convert amount to float
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount parameter", http.StatusBadRequest)
		return
	}

	// Settle billing for request
	if err := h.settleBilling(r.Context(), apiKey, amount); err != nil {
		http.Error(w, fmt.Sprintf("Billing failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Billing for API key %s settled for %.2f credits", apiKey, amount)))
}

// Deduct credits from API key account
func (h *BillingHandler) settleBilling(ctx context.Context, key string, amount float64) error {
	// Start a transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Retrieve current credit
	var currentCredit float64
	err = tx.QueryRowContext(ctx, "SELECT credit FROM users WHERE key = ?", key).Scan(&currentCredit)
	if err != nil {
		return fmt.Errorf("retrieving credits: %w", err)
	}

	// Check if sufficient credits are available
	if currentCredit < amount {
		return fmt.Errorf("insufficient credits")
	}

	// Update credit balance
	newCredit := currentCredit - amount
	_, err = tx.ExecContext(ctx, "UPDATE users SET credit = ? WHERE key = ?", newCredit, key)
	if err != nil {
		return fmt.Errorf("updating credit balance in database: %w", err)
	}

	// Update Redis with the new credit balance
	err = h.rdb.Set(ctx, key, newCredit, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("updating credit balance in Redis: %w", err)
	}

	// Commit the transaction
	return tx.Commit()
}

// Populates Redis with API keys and credits for testing
func seed(ctx context.Context, rdb *redis.Client, apiKey string, credits float64) {
	err := rdb.Set(ctx, apiKey, credits, 24*time.Hour).Err()
	if err != nil {
		log.Fatalf("Failed to seed Redis: %v", err)
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

	// Initialize Demo Table
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"key" TEXT,
		"credit" FLOAT,
		"created_at" DATETIME
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal(err)
	}

	// Populate table with demo user
	if _, err := db.Exec("INSERT INTO users (key, credit, created_at) VALUES (?, ?, ?)", DEMO_API_KEY, DEMO_API_CREDITS, time.Now()); err != nil {
		log.Fatal(err)
	}

	// Initialize Redis client using shared common logic
	rdb, err := common.NewRedisClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	// Seed Redis with demo API key and credits
	seed(context.Background(), rdb, DEMO_API_KEY, DEMO_API_CREDITS)

	billingHandler := &BillingHandler{db: db, rdb: rdb}

	http.Handle("/api/v1/billing", billingHandler)

	// Start the server
	fmt.Println("Go Billing is running on port 8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	_ "github.com/mattn/go-sqlite3"
)

const (
	KafkaBroker = "kafka:9092"
	KafkaTopic  = "coordinates"
	DBPath      = "/db/gps.db"
)

// CoordinateEvent represents a GPS coordinate event
type CoordinateEvent struct {
	UserID    string  `json:"user_id"`
	SessionID string  `json:"session_id,omitempty"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timestamp string  `json:"timestamp"`
}

// getEvents handles the HTTP endpoint for retrieving events
func getEvents(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")

		// Get query parameters
		userID := r.URL.Query().Get("user_id")
		limit := r.URL.Query().Get("limit")
		limitNum := 50 // default limit

		if limit != "" {
			if n, err := strconv.Atoi(limit); err == nil && n > 0 {
				limitNum = n
			}
		}

		// Build query
		query := `
			SELECT user_id, session_id, lat, lon, timestamp
			FROM coordinates
			WHERE 1=1
		`
		args := []interface{}{}

		if userID != "" {
			query += " AND user_id = ?"
			args = append(args, userID)
		}

		query += " ORDER BY timestamp DESC LIMIT ?"
		args = append(args, limitNum)

		// Execute query
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("Error querying database: %v\n", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Collect results
		events := []CoordinateEvent{}
		for rows.Next() {
			var event CoordinateEvent
			err := rows.Scan(
				&event.UserID,
				&event.SessionID,
				&event.Lat,
				&event.Lon,
				&event.Timestamp,
			)
			if err != nil {
				log.Printf("Error scanning row: %v\n", err)
				continue
			}
			events = append(events, event)
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}
}

func main() {
	// Initialize SQLite database
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS coordinates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			session_id TEXT,
			lat REAL NOT NULL,
			lon REAL NOT NULL,
			timestamp TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// Initialize Kafka consumer
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  KafkaBroker,
		"group.id":          "gps-consumer",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal("Failed to create consumer:", err)
	}
	defer c.Close()

	// Subscribe to topic
	err = c.SubscribeTopics([]string{KafkaTopic}, nil)
	if err != nil {
		log.Fatal("Failed to subscribe to topic:", err)
	}
	log.Printf("Subscribed to topic: %s\n", KafkaTopic)

	// Setup HTTP server
	http.HandleFunc("/events", getEvents(db))
	go func() {
		log.Printf("🚀 HTTP server running on :8082\n")
		if err := http.ListenAndServe(":8082", nil); err != nil {
			log.Fatal("HTTP server error:", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Prepare insert statement
	stmt, err := db.Prepare(`
		INSERT INTO coordinates (user_id, session_id, lat, lon, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal("Failed to prepare statement:", err)
	}
	defer stmt.Close()

	running := true
	for running {
		select {
		case sig := <-sigChan:
			log.Printf("Caught signal %v: terminating\n", sig)
			running = false
		default:
			ev := c.Poll(100) // 100ms timeout
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				var event CoordinateEvent
				if err := json.Unmarshal(e.Value, &event); err != nil {
					log.Printf("Error unmarshaling message: %v\n", err)
					continue
				}

				// Insert into SQLite
				_, err = stmt.Exec(
					event.UserID,
					event.SessionID,
					event.Lat,
					event.Lon,
					event.Timestamp,
				)
				if err != nil {
					log.Printf("Error inserting into database: %v\n", err)
					continue
				}

				log.Printf("Stored event: UserID=%s, Lat=%f, Lon=%f\n", event.UserID, event.Lat, event.Lon)

			case kafka.Error:
				log.Printf("Kafka error: %v\n", e)
			}
		}
	}
}

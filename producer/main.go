package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	KafkaBroker = "localhost:9094" // External broker address
)

// User represents a user with their base location
type User struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Base Location `json:"base"`
}

// Location represents a geographical point
type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// CoordinateEvent represents a GPS coordinate event
type CoordinateEvent struct {
	UserID    string  `json:"user_id"`
	SessionID string  `json:"session_id"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timestamp string  `json:"timestamp"`
}

// LocationEvent represents a location update event
type LocationEvent struct {
	UserID    string  `json:"user_id"`
	SessionID string  `json:"session_id"`
	Location  string  `json:"location"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timestamp string  `json:"timestamp"`
}

// Global variables
var (
	producer *kafka.Producer
	users    = []User{
		{ID: "Ashish", Name: "Ashish", Base: Location{Lat: 40.7128, Lon: -74.0060}},    // NYC
		{ID: "Saranya", Name: "Saranya", Base: Location{Lat: 34.0522, Lon: -118.2437}}, // LA
		{ID: "Cookie", Name: "Cookie", Base: Location{Lat: 51.5074, Lon: -0.1278}},     // London
	}
	locations = []string{"Park", "Trailhead", "Downtown", "Beach"}
)

// generateCoordinate simulates small random movement around base location
func generateCoordinate(base Location) Location {
	return Location{
		Lat: base.Lat + rand.Float64()*0.02 - 0.01, // +/- 0.01 degrees
		Lon: base.Lon + rand.Float64()*0.02 - 0.01,
	}
}

// deliveryReport handles delivery reports from Kafka producer
func deliveryReport(producer *kafka.Producer) {
	for e := range producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Printf("❌ Delivery failed for record: %v\n", ev.TopicPartition.Error)
			} else {
				log.Printf("✅ Message produced to %v [%d] @ offset %v\n",
					*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
			}
		case kafka.Error:
			log.Printf("❌ Kafka error: %v\n", ev)
		default:
			log.Printf("ℹ️ Ignored event: %s\n", ev)
		}
	}
}

// generateEvents continuously generates GPS events
func generateEvents(producer *kafka.Producer, done chan bool) {
	log.Printf("🌍 Starting GPS event generation for %d users\n", len(users))
	for {
		select {
		case <-done:
			log.Println("🛑 Stopping GPS event generation")
			return
		default:
			for _, user := range users {
				// Generate new coordinate
				pos := generateCoordinate(user.Base)
				now := time.Now().UTC().Format(time.RFC3339)

				// Create coordinate event
				coordEvent := CoordinateEvent{
					UserID:    user.ID,
					SessionID: "s1",
					Lat:       pos.Lat,
					Lon:       pos.Lon,
					Timestamp: now,
				}

				// Serialize to JSON
				coordData, err := json.Marshal(coordEvent)
				if err != nil {
					log.Printf("Error marshaling coordinate event: %v\n", err)
					continue
				}

				// Produce coordinate event
				err = producer.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &[]string{"coordinates"}[0], Partition: kafka.PartitionAny},
					Key:            []byte(user.ID),
					Value:          coordData,
				}, nil)

				if err != nil {
					log.Printf("Error producing coordinate event: %v\n", err)
				}

				// Occasionally emit a location event (10% chance)
				if rand.Float64() < 0.1 {
					locEvent := LocationEvent{
						UserID:    user.ID,
						SessionID: "s1",
						Location:  locations[rand.Intn(len(locations))],
						Lat:       pos.Lat,
						Lon:       pos.Lon,
						Timestamp: now,
					}

					// Serialize to JSON
					locData, err := json.Marshal(locEvent)
					if err != nil {
						log.Printf("Error marshaling location event: %v\n", err)
						continue
					}

					// Produce location event
					err = producer.Produce(&kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &[]string{"locations"}[0], Partition: kafka.PartitionAny},
						Key:            []byte(user.ID),
						Value:          locData,
					}, nil)

					if err != nil {
						log.Printf("Error producing location event: %v\n", err)
					}
				}

				// Trigger delivery report callbacks
				producer.Flush(0)
			}

			// Wait before next iteration
			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	log.Println("🚀 Starting GPS event producer...")

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Initialize Kafka producer
	var err error
	producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": KafkaBroker,
	})
	if err != nil {
		log.Fatal("Failed to create producer:", err)
	}
	defer producer.Close()

	// Start delivery report handler
	go deliveryReport(producer)

	// Channel to signal goroutine to stop
	done := make(chan bool)

	// Start GPS event generator
	go generateEvents(producer, done)

	// Setup HTTP endpoints
	http.HandleFunc("/produce", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		var event CoordinateEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Set timestamp if not provided
		if event.Timestamp == "" {
			event.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		// Serialize event
		data, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "Error serializing event", http.StatusInternalServerError)
			return
		}

		// Produce to Kafka
		err = producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &[]string{"coordinates"}[0], Partition: kafka.PartitionAny},
			Key:            []byte(event.UserID),
			Value:          data,
		}, nil)

		if err != nil {
			http.Error(w, "Error producing message", http.StatusInternalServerError)
			return
		}

		producer.Flush(1000)
		w.Write([]byte("ok"))
	})

	// Start HTTP server
	go func() {
		log.Printf("🚀 HTTP server running on :8081\n")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatal("HTTP server error:", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🚴 GPS generator started... (Ctrl+C to stop)")
	<-sigChan
	log.Println("\n📥 Shutting down...")

	// Stop GPS generator
	done <- true

	// Flush any remaining messages
	producer.Flush(5000)
}

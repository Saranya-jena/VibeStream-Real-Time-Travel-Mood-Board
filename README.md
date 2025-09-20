# Garmin Mock

A real-time GPS tracking simulation system built with Apache Kafka, Go, and React. It simulates Garmin-like GPS coordinate streaming for multiple users and visualizes their movements in real-time.

## Architecture

### Components

1. **Producer (Go)**
   - Simulates GPS coordinates for multiple users
   - Generates random movements around base locations
   - Publishes events to Kafka topics
   - Base locations:
     - NYC
     - LA
     - London

2. **Consumer (Go)**
   - Subscribes to Kafka topics
   - Stores coordinates in SQLite database
   - Provides REST API for querying historical data
   - Endpoints:
     - GET `/events` - Fetch all events
     - GET `/events?user_id=<id>` - Fetch user-specific events

3. **Frontend (React)**
   - Real-time display of user locations
   - User filtering capability
   - Auto-refreshing event list
   - Clean, modern UI

4. **Kafka**
   - Message broker for event streaming
   - Topics:
     - `coordinates` - GPS coordinate events
     - `locations` - Location update events

## Setup

### Prerequisites
- Docker and Docker Compose
- Go 1.21+
- Node.js 16+
- npm

### Running the Application

1. Start the Kafka infrastructure and consumer:
   ```bash
   docker compose up --build
   ```

2. Run the producer:
   ```bash
   cd producer
   go run main.go
   ```

3. Start the frontend:
   ```bash
   cd frontend
   npm install
   npm start
   ```

### Ports
- Kafka: 9092 (internal), 9094 (external)
- Kafka UI: 8080
- Consumer API: 8082
- Frontend: 3000

## Data Flow

1. Producer generates simulated GPS coordinates
2. Events are published to Kafka topics
3. Consumer processes events and stores them in SQLite
4. Frontend polls the consumer API every 5 seconds
5. UI updates with the latest coordinates

## Features

- Real-time GPS coordinate simulation
- Persistent storage in SQLite
- User-based filtering
- Automatic UI updates
- Clean, responsive design

## Project Structure

```
├── consumer/           # Go consumer service
│   ├── main.go        # Consumer implementation
│   └── Dockerfile     # Consumer container setup
├── producer/          # Go producer service
│   └── main.go        # Producer implementation
├── frontend/          # React frontend
│   ├── src/           # React components
│   └── package.json   # Frontend dependencies
├── db/                # SQLite database directory
└── docker-compose.yml # Service orchestration
```

## Contributing

Feel free to submit issues, fork the repository, and create pull requests for any improvements.

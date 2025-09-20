#!/bin/bash

# List of topics to create
TOPICS=("coordinates" "locations" "users")

# Kafka container name (matches docker-compose.yml)
CONTAINER_NAME="kafka"

# Kafka bootstrap server
BOOTSTRAP="127.0.0.1:9092"

# Function to check if Kafka is ready
check_kafka() {
  docker exec -it $CONTAINER_NAME kafka-topics --list --bootstrap-server $BOOTSTRAP >/dev/null 2>&1
}

echo "⏳ Waiting for Kafka to be ready on $BOOTSTRAP..."

# Retry loop (up to 30 tries with 2s pause = ~60s max wait)
RETRIES=30
for i in $(seq 1 $RETRIES); do
  if check_kafka; then
    echo "✅ Kafka is ready!"
    break
  fi
  echo "Kafka not ready yet... ($i/$RETRIES)"
  sleep 2
done

# If still not ready, exit with error
if ! check_kafka; then
  echo "❌ Kafka did not become ready after $RETRIES attempts."
  exit 1
fi

# Create topics
for topic in "${TOPICS[@]}"; do
  echo "📌 Creating topic: $topic"
  docker exec -it $CONTAINER_NAME kafka-topics \
    --create \
    --if-not-exists \
    --topic "$topic" \
    --bootstrap-server $BOOTSTRAP
done

# List topics
echo "📋 Current topics:"
docker exec -it $CONTAINER_NAME kafka-topics --list --bootstrap-server $BOOTSTRAP

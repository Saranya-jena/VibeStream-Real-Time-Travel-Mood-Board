#!/bin/bash
set -e

echo "Starting to create topics..."

KAFKA_CMD="docker exec kafka /opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server kafka:9092"

# Create users topic
$KAFKA_CMD --create --if-not-exists \
  --replication-factor 1 \
  --partitions 1 \
  --topic users

# Create locations topic
$KAFKA_CMD --create --if-not-exists \
  --replication-factor 1 \
  --partitions 1 \
  --topic locations

# Create coordinates topic
$KAFKA_CMD --create --if-not-exists \
  --replication-factor 1 \
  --partitions 1 \
  --topic coordinates

# List all topics
echo "\nListing all topics:"
$KAFKA_CMD --list

echo "\nKafka topics created successfully."
# Kafka Integration — Implementation Summary

## What was implemented

### 1. `pkg/kafka/events.go` — Shared types and constants
- 4 topic constants: `sports.events`, `sports.notifications`, `sports.analytics`, `sports.dlq`
- 3 event type constants: `TrainingBooked`, `SubscriptionExpired`, `TrainingCancelled`
- Unified `Event` struct with `EventID`, `EventType`, `EntityID`, `Timestamp`, `Source`, `Payload`, `Version`, `Metadata`
- `EventMetadata` with `CorrelationID`, `ProducedAt`, `Env`
- 3 payload structs: `TrainingBookedPayload`, `SubscriptionExpiredPayload`, `TrainingCancelledPayload`
- `NewUUID()` helper using `crypto/rand`

### 2. `cmd/kafka-producer/main.go` — Event Producer
- Produces 3 event types in a random loop
- Message key = `entityId` (enables partition affinity per client)
- Kafka headers: `eventType`, `source`, `version`
- Full `Event` struct with metadata on every message
- `-count` flag (default 30, 0 = infinite)
- Random sleep 200–1000ms between events
- Uses `kafka.Hash{}` balancer for key-based routing

### 3. `cmd/kafka-consumer/main.go` — Two Consumer Groups
**`notifications-group`:**
- Consumes only `TrainingBooked` events
- Forwards notifications to `sports.notifications` topic
- 3-attempt retry with backoff on processing failure
- Sends to DLQ on `max_retries_exceeded` or unmarshal errors
- Manual commit via `FetchMessage` + `CommitMessages`

**`audit-group`:**
- Consumes ALL event types
- Logs full audit trail: partition, offset, eventType, entityId, source, timestamp
- Sends malformed messages to DLQ
- Manual commit

Both groups share `publishToDLQ()` helper that appends `dlq-reason` and `dlq-time` headers.

### 4. `cmd/kafka-streams/main.go` — Stream Processor
Consumer group `streams-group` with three processing stages:

**Transformation:** Every event is enriched into an `EnrichedEvent` with `processedAt`, `processingMs`, `streamGroup`, and a normalized flat map.

**Aggregation:** Counts events per type (`countByType` map), flushed every 10 seconds to `sports.analytics`.

**Windowed calculation:** Tumbling window of 60 seconds — counts events per `entityId` within the current window (`windowCounts`). Window resets automatically when the 60s duration elapses.

Flush goroutine publishes `AggregateResult` JSON to `sports.analytics` with `window-aggregate` header.

### 5. `configs/kafka/mongo-sink-connector.json` — Kafka Connect Sink
- Reads from `sports.events` topic
- Writes to MongoDB `sport_club.kafka_events` collection
- Uses `InsertOneDefaultStrategy` and UUID document ID strategy
- JSON value converter without schemas

### 6. `configs/kafka/mongo-source-connector.json` — Kafka Connect Source
- Watches MongoDB `sport_club.users` collection via Change Streams
- Captures `insert`, `update`, `replace` operations
- Publishes to `sports.mongo.*` topics
- Publishes full document only

### 7. `docker/docker-compose.yml` — Kafka Stack Services
Added 4 new services:
- **zookeeper** (confluentinc/cp-zookeeper:7.5.0) on port 2181
- **kafka** (confluentinc/cp-kafka:7.5.0) on port 9092, auto-creates topics
- **kafka-connect** (confluentinc/cp-kafka-connect:7.5.0) on port 8083, installs MongoDB connector plugin on startup
- **kafka-ui** (provectuslabs/kafka-ui) on port 8080 with connect integration

### 8. `READMEkafka.md` — Quick Start Guide
Step-by-step instructions to start the stack, run producer/consumers/streams, view Kafka UI, and register Connect connectors.

## Architecture

```
Producer --> sports.events --> notifications-group --> sports.notifications
                          --> audit-group (logs all)
                          --> streams-group --> sports.analytics (aggregates)
                          --> (errors) --> sports.dlq
```

## Dependencies
- `github.com/segmentio/kafka-go` — pure Go Kafka client

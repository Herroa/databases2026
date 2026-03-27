QUICK START — Kafka

## 1. Start Kafka stack

cd docker
docker compose up -d zookeeper kafka kafka-ui

wait ~15 seconds for kafka to be ready

## 2. Run producer (generates 30 events)

go run ../cmd/kafka-producer/main.go -count 30

## 3. Run consumers (open 2 terminals)

# Terminal 1 — notifications group
go run ../cmd/kafka-consumer/main.go -group notifications-group

# Terminal 2 — audit group
go run ../cmd/kafka-consumer/main.go -group audit-group

## 4. Run Kafka Streams processor (open another terminal)

go run ../cmd/kafka-streams/main.go

## 5. View in Kafka UI

http://localhost:8080

Topics: sports.events | sports.notifications | sports.analytics | sports.dlq

## 6. Kafka Connect (optional, requires MongoDB running)

cd docker
docker compose up -d kafka-connect

# wait ~60s for connector plugins to install, then:

# Register sink connector (events -> MongoDB):
curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d @../configs/kafka/mongo-sink-connector.json

# Register source connector (MongoDB users -> Kafka):
curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d @../configs/kafka/mongo-source-connector.json

# Check connector status:
curl http://localhost:8083/connectors/sports-events-mongo-sink/status

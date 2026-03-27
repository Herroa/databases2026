package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	kafkapkg "databases2026/pkg/kafka"
	"github.com/segmentio/kafka-go"
)

const brokerAddr = "localhost:9092"

func main() {
	count := flag.Int("count", 30, "number of events to produce (0 = infinite)")
	flag.Parse()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokerAddr),
		Topic:        kafkapkg.TopicEvents,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
	}
	defer writer.Close()

	sports := []string{"Yoga", "Boxing", "Crossfit", "Pilates", "Swimming"}
	reasons := []string{"coach_sick", "room_unavailable", "client_request", "schedule_conflict"}
	subTypes := []string{"monthly", "annual", "trial"}

	clients := make([]string, 10)
	for i := range clients {
		clients[i] = fmt.Sprintf("client-%d", i+1)
	}

	produced := 0
	for *count == 0 || produced < *count {
		var event kafkapkg.Event
		clientID := clients[rand.Intn(len(clients))]

		switch rand.Intn(3) {
		case 0:
			payload := kafkapkg.TrainingBookedPayload{
				ClientID:    clientID,
				TrainingID:  kafkapkg.NewUUID(),
				CoachID:     fmt.Sprintf("coach-%d", rand.Intn(5)+1),
				RoomID:      fmt.Sprintf("room-%d", rand.Intn(3)+1),
				Sport:       sports[rand.Intn(len(sports))],
				ScheduledAt: time.Now().Add(time.Duration(rand.Intn(72)) * time.Hour),
			}
			raw, _ := json.Marshal(payload)
			event = kafkapkg.Event{
				EventID:   kafkapkg.NewUUID(),
				EventType: kafkapkg.EventTrainingBooked,
				EntityID:  clientID,
				Timestamp: time.Now(),
				Source:    "sports-center-app",
				Version:   "1.0",
				Payload:   raw,
				Metadata: kafkapkg.EventMetadata{
					CorrelationID: kafkapkg.NewUUID(),
					ProducedAt:    time.Now().UTC().Format(time.RFC3339),
					Env:           "dev",
				},
			}

		case 1:
			payload := kafkapkg.SubscriptionExpiredPayload{
				ClientID:         clientID,
				SubscriptionType: subTypes[rand.Intn(len(subTypes))],
				ExpiredAt:        time.Now().Add(-time.Duration(rand.Intn(5)) * 24 * time.Hour),
				DaysOverdue:      rand.Intn(30) + 1,
			}
			raw, _ := json.Marshal(payload)
			event = kafkapkg.Event{
				EventID:   kafkapkg.NewUUID(),
				EventType: kafkapkg.EventSubscriptionExpired,
				EntityID:  clientID,
				Timestamp: time.Now(),
				Source:    "subscription-service",
				Version:   "1.0",
				Payload:   raw,
				Metadata: kafkapkg.EventMetadata{
					CorrelationID: kafkapkg.NewUUID(),
					ProducedAt:    time.Now().UTC().Format(time.RFC3339),
					Env:           "dev",
				},
			}

		case 2:
			trainingID := kafkapkg.NewUUID()
			payload := kafkapkg.TrainingCancelledPayload{
				TrainingID:  trainingID,
				ClientID:    clientID,
				Reason:      reasons[rand.Intn(len(reasons))],
				CancelledAt: time.Now(),
			}
			raw, _ := json.Marshal(payload)
			event = kafkapkg.Event{
				EventID:   kafkapkg.NewUUID(),
				EventType: kafkapkg.EventTrainingCancelled,
				EntityID:  trainingID,
				Timestamp: time.Now(),
				Source:    "sports-center-app",
				Version:   "1.0",
				Payload:   raw,
				Metadata: kafkapkg.EventMetadata{
					CorrelationID: kafkapkg.NewUUID(),
					ProducedAt:    time.Now().UTC().Format(time.RFC3339),
					Env:           "dev",
				},
			}
		}

		value, err := json.Marshal(event)
		if err != nil {
			log.Printf("marshal error: %v", err)
			continue
		}

		msg := kafka.Message{
			Key:   []byte(event.EntityID),
			Value: value,
			Headers: []kafka.Header{
				{Key: "eventType", Value: []byte(event.EventType)},
				{Key: "source", Value: []byte(event.Source)},
				{Key: "version", Value: []byte(event.Version)},
			},
		}

		if err := writer.WriteMessages(context.Background(), msg); err != nil {
			log.Printf("write error: %v", err)
		} else {
			log.Printf("produced [%s] entityId=%s", event.EventType, event.EntityID)
		}

		produced++
		time.Sleep(time.Duration(200+rand.Intn(800)) * time.Millisecond)
	}

	fmt.Printf("Produced %d events\n", produced)
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafkapkg "databases2026/pkg/kafka"
	"github.com/segmentio/kafka-go"
)

const brokerAddr = "localhost:9092"

func main() {
	group := flag.String("group", "notifications-group", "consumer group: notifications-group | audit-group")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	switch *group {
	case "notifications-group":
		runNotificationsConsumer(ctx)
	case "audit-group":
		runAuditConsumer(ctx)
	default:
		fmt.Println("unknown group, use: notifications-group | audit-group")
		os.Exit(1)
	}
}

func newReader(groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{brokerAddr},
		Topic:          kafkapkg.TopicEvents,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0, // manual commit
		StartOffset:    kafka.FirstOffset,
	})
}

func newDLQWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokerAddr),
		Topic:    kafkapkg.TopicDLQ,
		Balancer: &kafka.LeastBytes{},
	}
}

func publishToDLQ(ctx context.Context, w *kafka.Writer, original kafka.Message, reason string) {
	msg := kafka.Message{
		Key:   original.Key,
		Value: original.Value,
		Headers: append(original.Headers,
			kafka.Header{Key: "dlq-reason", Value: []byte(reason)},
			kafka.Header{Key: "dlq-time", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
		),
	}
	if err := w.WriteMessages(ctx, msg); err != nil {
		log.Printf("DLQ write error: %v", err)
	} else {
		log.Printf("-> DLQ: %s", reason)
	}
}

// Consumer group 1: notifications
func runNotificationsConsumer(ctx context.Context) {
	reader := newReader("notifications-group")
	defer reader.Close()

	dlq := newDLQWriter()
	defer dlq.Close()

	notifWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokerAddr),
		Topic:    kafkapkg.TopicNotifications,
		Balancer: &kafka.LeastBytes{},
	}
	defer notifWriter.Close()

	log.Println("notifications-group consumer started")

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("fetch error: %v", err)
			continue
		}

		var event kafkapkg.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("unmarshal error: %v", err)
			publishToDLQ(ctx, dlq, msg, fmt.Sprintf("unmarshal_error: %v", err))
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Only process TrainingBooked
		if event.EventType != kafkapkg.EventTrainingBooked {
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Simulate processing with retry
		processed := false
		for attempt := 1; attempt <= 3; attempt++ {
			if err := processNotification(ctx, event, notifWriter); err != nil {
				log.Printf("attempt %d failed: %v", attempt, err)
				time.Sleep(time.Duration(attempt*100) * time.Millisecond)
				continue
			}
			processed = true
			break
		}

		if !processed {
			publishToDLQ(ctx, dlq, msg, "max_retries_exceeded")
		}

		// Manual commit after processing
		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit error: %v", err)
		}
	}
}

func processNotification(ctx context.Context, event kafkapkg.Event, w *kafka.Writer) error {
	var payload kafkapkg.TrainingBookedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	notification := map[string]any{
		"notificationId": kafkapkg.NewUUID(),
		"clientId":       payload.ClientID,
		"message":        fmt.Sprintf("Training booked: %s on %s", payload.Sport, payload.ScheduledAt.Format("2006-01-02 15:04")),
		"sentAt":         time.Now().UTC().Format(time.RFC3339),
		"sourceEventId":  event.EventID,
	}

	raw, _ := json.Marshal(notification)
	msg := kafka.Message{
		Key:   []byte(payload.ClientID),
		Value: raw,
	}

	if err := w.WriteMessages(ctx, msg); err != nil {
		return err
	}

	log.Printf("notification sent to client %s: %s", payload.ClientID, payload.Sport)
	return nil
}

// Consumer group 2: audit
func runAuditConsumer(ctx context.Context) {
	reader := newReader("audit-group")
	defer reader.Close()

	dlq := newDLQWriter()
	defer dlq.Close()

	log.Println("audit-group consumer started")

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("fetch error: %v", err)
			continue
		}

		var event kafkapkg.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			publishToDLQ(ctx, dlq, msg, fmt.Sprintf("unmarshal_error: %v", err))
			reader.CommitMessages(ctx, msg)
			continue
		}

		log.Printf("AUDIT [partition=%d offset=%d] eventType=%s entityId=%s source=%s ts=%s",
			msg.Partition, msg.Offset,
			event.EventType, event.EntityID, event.Source,
			event.Timestamp.Format(time.RFC3339))

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit error: %v", err)
		}
	}
}

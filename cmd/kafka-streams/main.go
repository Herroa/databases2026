package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	kafkapkg "databases2026/pkg/kafka"
	"github.com/segmentio/kafka-go"
)

const brokerAddr = "localhost:9092"

type AggregateResult struct {
	WindowStart  time.Time      `json:"windowStart"`
	WindowEnd    time.Time      `json:"windowEnd"`
	CountByType  map[string]int `json:"countByType"`
	WindowCounts map[string]int `json:"windowEntityCounts"` // entityId -> count in window
	ProcessedAt  string         `json:"processedAt"`
}

type EnrichedEvent struct {
	Original     kafkapkg.Event `json:"original"`
	ProcessedAt  string         `json:"processedAt"`
	ProcessingMs int64          `json:"processingMs"`
	StreamGroup  string         `json:"streamGroup"`
	Normalized   map[string]any `json:"normalized"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; cancel() }()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{brokerAddr},
		Topic:          kafkapkg.TopicEvents,
		GroupID:        "streams-group",
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	})
	defer reader.Close()

	analyticsWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokerAddr),
		Topic:    kafkapkg.TopicAnalytics,
		Balancer: &kafka.LeastBytes{},
	}
	defer analyticsWriter.Close()

	var mu sync.Mutex
	countByType := make(map[string]int)

	windowDuration := 60 * time.Second
	windowStart := time.Now()
	windowCounts := make(map[string]int) // entityId -> count in current window

	// Ticker to flush aggregations
	flushTicker := time.NewTicker(10 * time.Second)
	defer flushTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-flushTicker.C:
				mu.Lock()
				result := AggregateResult{
					WindowStart:  windowStart,
					WindowEnd:    t,
					CountByType:  copyMap(countByType),
					WindowCounts: copyMap(windowCounts),
					ProcessedAt:  time.Now().UTC().Format(time.RFC3339),
				}

				// Reset window if expired
				if t.Sub(windowStart) >= windowDuration {
					windowCounts = make(map[string]int)
					windowStart = t
					log.Println("Window reset")
				}
				mu.Unlock()

				raw, _ := json.Marshal(result)
				msg := kafka.Message{
					Key:   []byte("aggregate"),
					Value: raw,
					Headers: []kafka.Header{
						{Key: "type", Value: []byte("window-aggregate")},
					},
				}
				if err := analyticsWriter.WriteMessages(context.Background(), msg); err != nil {
					log.Printf("analytics write error: %v", err)
				} else {
					log.Printf("analytics flushed: %v", result.CountByType)
				}
			}
		}
	}()

	log.Println("kafka-streams processor started")

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("fetch error: %v", err)
			continue
		}

		start := time.Now()

		var event kafkapkg.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("unmarshal error: %v", err)
			reader.CommitMessages(ctx, msg)
			continue
		}

		// === TRANSFORMATION: enrich event ===
		enriched := EnrichedEvent{
			Original:     event,
			ProcessedAt:  time.Now().UTC().Format(time.RFC3339),
			ProcessingMs: time.Since(start).Milliseconds(),
			StreamGroup:  "streams-group",
			Normalized: map[string]any{
				"type":     event.EventType,
				"entity":   event.EntityID,
				"source":   event.Source,
				"unixTime": event.Timestamp.Unix(),
			},
		}
		_ = enriched // used for logging/demonstration

		// === AGGREGATION: count by event type ===
		mu.Lock()
		countByType[event.EventType]++

		// === WINDOWED: count per entityId in current window ===
		windowCounts[event.EntityID]++
		mu.Unlock()

		log.Printf("streams [%s] entity=%s | totals=%v",
			event.EventType, event.EntityID, fmt.Sprintf("%v", countByType))

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit error: %v", err)
		}
	}
}

func copyMap(m map[string]int) map[string]int {
	cp := make(map[string]int, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

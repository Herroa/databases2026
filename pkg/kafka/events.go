package kafka

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// Topics
const (
	TopicEvents        = "sports.events"
	TopicNotifications = "sports.notifications"
	TopicAnalytics     = "sports.analytics"
	TopicDLQ           = "sports.dlq"
)

// Event types
const (
	EventTrainingBooked      = "TrainingBooked"
	EventSubscriptionExpired = "SubscriptionExpired"
	EventTrainingCancelled   = "TrainingCancelled"
)

type EventMetadata struct {
	CorrelationID string `json:"correlationId"`
	ProducedAt    string `json:"producedAt"`
	Env           string `json:"env"`
}

type Event struct {
	EventID   string          `json:"eventId"`
	EventType string          `json:"eventType"`
	EntityID  string          `json:"entityId"`
	Timestamp time.Time       `json:"timestamp"`
	Source    string          `json:"source"`
	Payload   json.RawMessage `json:"payload"`
	Version   string          `json:"version"`
	Metadata  EventMetadata   `json:"metadata"`
}

type TrainingBookedPayload struct {
	ClientID    string    `json:"clientId"`
	TrainingID  string    `json:"trainingId"`
	CoachID     string    `json:"coachId"`
	RoomID      string    `json:"roomId"`
	Sport       string    `json:"sport"`
	ScheduledAt time.Time `json:"scheduledAt"`
}

type SubscriptionExpiredPayload struct {
	ClientID         string    `json:"clientId"`
	SubscriptionType string    `json:"subscriptionType"`
	ExpiredAt        time.Time `json:"expiredAt"`
	DaysOverdue      int       `json:"daysOverdue"`
}

type TrainingCancelledPayload struct {
	TrainingID  string    `json:"trainingId"`
	ClientID    string    `json:"clientId"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelledAt"`
}

func NewUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

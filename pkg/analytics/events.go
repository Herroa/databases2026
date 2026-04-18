package analytics

type AnalyticsEvent struct {
    EventID        string    `ch:"event_id"`            // UUID
    EventType      string    `ch:"event_type"`          // TrainingBooked | TrainingCancelled | SubscriptionPurchased
    EntityType     string    `ch:"entity_type"`         // training | subscription
    EntityID       string    `ch:"entity_id"`           // ID сущности (тренировка/подписка)
    EventTimestamp time.Time `ch:"event_timestamp"`     // Время события (event time)
    IngestedAt     time.Time `ch:"ingested_at"`         // Время загрузки в CH (processing time)
    ClientID       string    `ch:"client_id"`           // ID клиента (сквозной ключ)
    CoachID        string    `ch:"coach_id,omitempty"`  // Только для Training*
    RoomID         string    `ch:"room_id,omitempty"`   // Только для Training*
    SportType      string    `ch:"sport_type,omitempty"`// Yoga, Boxing, Crossfit...
    Price          float64   `ch:"price"`               // Стоимость (0 для отмен)
    RefundAmount   float64   `ch:"refund_amount"`       // Сумма возврата
    Currency       string    `ch:"currency"`            // RUB, USD...
    ScheduledAt    time.Time `ch:"scheduled_at,omitempty"` // Запланированное время тренировки
    CancelledAt    time.Time `ch:"cancelled_at,omitempty"` // Время отмены
    HoursBefore    int32     `ch:"hours_before_training"`  // За сколько часов отменили
    Source         string    `ch:"source"`              // sports-center-app, mobile-app, web
    PromoCode      string    `ch:"promo_code,omitempty"`// Промокод
    PaymentMethod  string    `ch:"payment_method,omitempty"` // card, cash, subscription
    CancelReason   string    `ch:"cancel_reason,omitempty"`// coach_sick, client_request...   
    Version        uint32    `ch:"version"`   
    IsDeleted      uint8     `ch:"is_deleted"`
}
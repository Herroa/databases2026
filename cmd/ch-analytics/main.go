// cmd/ch-analytics/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

const (
	batchSize    = 5000
	totalRecords = 50000
)

type Event struct {
	EventID        string
	EventType      string
	EntityType     string
	EntityID       string
	EventTimestamp time.Time
	ClientID       string
	CoachID        *string
	RoomID         *string
	SportType      *string
	Price          float64
	RefundAmount   float64
	Currency       string
	ScheduledAt    *time.Time
	CancelledAt    *time.Time
	HoursBefore    *int32
	Source         string
	PromoCode      *string
	PaymentMethod  *string
	CancelReason   *string
	Version        uint32
	IsDeleted      uint8
}

var (
	sports         = []string{"Yoga", "Boxing", "Crossfit", "Pilates", "Swimming", "TRX", "Zumba"}
	coaches        = make([]string, 15)
	rooms          = []string{"room-1", "room-2", "room-3", "hall-main", "pool"}
	clients        = make([]string, 500)
	promoCodes     = []string{"SUMMER20", "NEWCLIENT", "LOYAL10", "", "", ""}
	cancelReasons  = []string{"coach_sick", "client_request", "schedule_conflict", "illness", ""}
	paymentMethods = []string{"card", "cash", "subscription", "corporate"}
	sources        = []string{"sports-center-app", "mobile-app", "web-portal"}
)

func init() {
	rand.Seed(time.Now().UnixNano())
	for i := range coaches {
		coaches[i] = fmt.Sprintf("coach-%02d", i+1)
	}
	for i := range clients {
		clients[i] = fmt.Sprintf("client-%04d", i+1)
	}
}

func genEvent(base time.Time) *Event {
	hour := rand.Intn(24)
	if rand.Float32() < 0.4 {
		hour = rand.Intn(3) + 7
	} else if rand.Float32() < 0.8 {
		hour = rand.Intn(4) + 18
	}
	ts := base.Add(time.Duration(rand.Intn(30)*24+hour) * time.Hour)

	var et string
	r := rand.Float32()
	if r < 0.5 {
		et = "TrainingBooked"
	} else if r < 0.8 {
		et = "SubscriptionPurchased"
	} else {
		et = "TrainingCancelled"
	}

	e := &Event{
		EventID:        uuid.New().String(),
		EventType:      et,
		EventTimestamp: ts,
		ClientID:       clients[rand.Intn(len(clients))],
		Source:         sources[rand.Intn(len(sources))],
		Currency:       "RUB",
		Version:        1,
	}

	switch et {
	case "TrainingBooked":
		e.EntityType = "training"
		e.EntityID = uuid.New().String()
		e.SportType = &sports[rand.Intn(len(sports))]
		e.CoachID = &coaches[rand.Intn(len(coaches))]
		e.RoomID = &rooms[rand.Intn(len(rooms))]
		e.Price = 500 + rand.Float64()*2500
		e.ScheduledAt = &ts
		pm := paymentMethods[rand.Intn(len(paymentMethods))]
		e.PaymentMethod = &pm
		if rand.Float32() < 0.3 {
			pc := promoCodes[rand.Intn(3)]
			e.PromoCode = &pc
		}
	case "TrainingCancelled":
		e.EntityType = "training"
		e.EntityID = uuid.New().String()
		e.SportType = &sports[rand.Intn(len(sports))]
		e.Price = 500 + rand.Float64()*2500
		e.RefundAmount = e.Price * rand.Float64()
		reason := cancelReasons[rand.Intn(len(cancelReasons))]
		e.CancelReason = &reason
		hb := int32(rand.Intn(168))
		e.HoursBefore = &hb
		e.ScheduledAt = &ts
		e.CancelledAt = &ts
	case "SubscriptionPurchased":
		e.EntityType = "subscription"
		e.EntityID = uuid.New().String()
		types := []string{"monthly", "quarterly", "annual", "trial"}
		prices := map[string]float64{"monthly": 2990, "quarterly": 7990, "annual": 24990, "trial": 0}
		t := types[rand.Intn(len(types))]
		e.Price = prices[t]
		if rand.Float32() < 0.25 {
			pc := promoCodes[rand.Intn(3)]
			e.PromoCode = &pc
			e.Price *= 0.9
		}
		pm := paymentMethods[rand.Intn(len(paymentMethods))]
		e.PaymentMethod = &pm
	}
	return e
}

const createTable = `
CREATE TABLE IF NOT EXISTS sports_center.analytics_events (
    event_id UUID,
    event_type LowCardinality(String),
    entity_type LowCardinality(String),
    entity_id String,
    event_timestamp DateTime('UTC'),
    client_id String,
    coach_id Nullable(String),
    room_id Nullable(String),
    sport_type LowCardinality(Nullable(String)),
    price Float64,
    refund_amount Float64,
    currency LowCardinality(String),
    scheduled_at Nullable(DateTime('UTC')),
    cancelled_at Nullable(DateTime('UTC')),
    hours_before Nullable(Int32),
    source LowCardinality(String),
    promo_code Nullable(String),
    payment_method LowCardinality(Nullable(String)),
    cancel_reason LowCardinality(Nullable(String)),
    version UInt32,
    is_deleted UInt8,
    event_date Date MATERIALIZED toDate(event_timestamp),
    event_hour UInt8 MATERIALIZED toHour(event_timestamp)
) ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_type, client_id, event_timestamp, event_id)
TTL event_timestamp + INTERVAL 18 MONTH DELETE;
`

const createMetrics = `
CREATE TABLE IF NOT EXISTS sports_center.daily_client_metrics (
    metric_date Date,
    client_id String,
    trainings_booked UInt32,
    trainings_cancelled UInt32,
    subscriptions_purchased UInt32,
    total_spent Float64,
    total_refunded Float64,
    avg_order_value Float64,
    first_event_time DateTime('UTC'),
    last_event_time DateTime('UTC'),
    preferred_sport LowCardinality(Nullable(String)),
    preferred_payment LowCardinality(Nullable(String)),
    version UInt64
) ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(metric_date)
ORDER BY (metric_date, client_id);
`

const createMV = `
CREATE MATERIALIZED VIEW IF NOT EXISTS sports_center.mv_daily_client_metrics
TO sports_center.daily_client_metrics AS
SELECT
    toDate(event_timestamp) as metric_date,
    client_id,
    countIf(event_type = 'TrainingBooked') as trainings_booked,
    countIf(event_type = 'TrainingCancelled') as trainings_cancelled,
    countIf(event_type = 'SubscriptionPurchased') as subscriptions_purchased,
    sumIf(price, event_type IN ('TrainingBooked', 'SubscriptionPurchased')) as total_spent,
    sumIf(refund_amount, event_type = 'TrainingCancelled') as total_refunded,
    round(avgIf(price, price > 0), 2) as avg_order_value,
    min(event_timestamp) as first_event_time,
    max(event_timestamp) as last_event_time,
    argMax(sport_type, sc) as preferred_sport,
    argMax(payment_method, pc) as preferred_payment,
    uniqCombined64(event_id) as version
FROM (
    SELECT *,
        count() OVER (PARTITION BY toDate(event_timestamp), client_id, sport_type) as sc,
        count() OVER (PARTITION BY toDate(event_timestamp), client_id, payment_method) as pc
    FROM sports_center.analytics_events
    WHERE client_id != ''
)
GROUP BY metric_date, client_id;
`

type Query struct{ Name, Desc, SQL string }

var queries = []Query{
	{"room_utilization", "Загрузка залов", `
		SELECT event_hour as hour, sport_type, room_id, count() as bookings, uniq(client_id) as clients, round(avg(price),2) as avg_price
		FROM sports_center.analytics_events
		WHERE event_type='TrainingBooked' AND event_date >= today()-7 AND sport_type IS NOT NULL
		GROUP BY hour, sport_type, room_id ORDER BY hour, bookings DESC LIMIT 10`},
	{"coach_cancellation_rate", "Отмены по тренерам", `
		SELECT coach_id, sport_type, countIf(event_type='TrainingBooked') as total, countIf(event_type='TrainingCancelled') as canc,
			round(canc*100/total,2) as rate_pct
		FROM sports_center.analytics_events WHERE event_date>=today()-30 AND coach_id IS NOT NULL
		GROUP BY coach_id, sport_type HAVING total>=5 ORDER BY rate_pct DESC LIMIT 10`},
	{"revenue_by_subscription", "Выручка по подпискам", `
		SELECT payment_method, count() as purchases, uniq(client_id) as buyers, sum(price) as revenue, round(avg(price),2) as avg_price
		FROM sports_center.analytics_events WHERE event_type='SubscriptionPurchased' AND event_date>=today()-90
		GROUP BY payment_method ORDER BY revenue DESC`},
	{"client_ltv", "LTV клиента", `
		WITH m AS (SELECT client_id, sumIf(price, event_type IN ('TrainingBooked','SubscriptionPurchased')) as spent,
			countIf(event_type='TrainingBooked') as bookings FROM sports_center.analytics_events
			WHERE event_date>=today()-180 GROUP BY client_id)
		SELECT round(avg(spent),2) as avg_ltv, median(spent) as med_ltv, countIf(spent>0) as active, round(avg(bookings),1) as avg_bookings FROM m`},
	{"promo_effectiveness", "Эффективность промо", `
		SELECT promo_code, count() as uses, uniq(client_id) as users, sum(price) as revenue, round(avg(price),2) as avg_order
		FROM sports_center.analytics_events WHERE event_type='SubscriptionPurchased' AND promo_code!='' AND event_date>=today()-90
		GROUP BY promo_code ORDER BY revenue DESC LIMIT 5`},
	{"sport_popularity", "Популярность направлений", `
		SELECT sport_type, countIf(event_type='TrainingBooked') as bookings, uniqIf(client_id, event_type='TrainingBooked') as clients,
			round(avg(price),2) as avg_price, round(countIf(event_type='TrainingCancelled')*100/nullIf(countIf(event_type IN ('TrainingBooked','TrainingCancelled')),0),2) as cancel_rate
		FROM sports_center.analytics_events WHERE sport_type IS NOT NULL AND event_date>=today()-60
		GROUP BY sport_type HAVING bookings>=20 ORDER BY bookings DESC LIMIT 10`},
	{"daily_revenue_trend", "Динамика выручки", `
		SELECT event_date, event_type, count() as cnt, sum(price) as revenue, round(avg(price),2) as avg_price
		FROM sports_center.analytics_events WHERE event_date>=today()-30 AND event_type IN ('TrainingBooked','SubscriptionPurchased')
		GROUP BY event_date, event_type ORDER BY event_date, event_type LIMIT 30`},
	{"churn_risk", "Риск оттока", `
		WITH a AS (SELECT client_id, max(event_timestamp) as last, countIf(event_type='TrainingCancelled') as canc,
			countIf(event_type='TrainingBooked') as book, dateDiff('day', max(event_timestamp), now()) as days
			FROM sports_center.analytics_events WHERE event_date>=today()-90 GROUP BY client_id)
		SELECT client_id, days, round(canc*100/nullIf(book+canc,0),2) as cancel_rate,
			CASE WHEN days>30 THEN 'HIGH' WHEN days>14 THEN 'MEDIUM' ELSE 'LOW' END as risk
		FROM a WHERE book+canc>=2 ORDER BY risk DESC, cancel_rate DESC LIMIT 10`},
}

func main() {
	ctx := context.Background()
	fmt.Println("🏋️ ClickHouse Analytics for Sports Center")
	fmt.Println("==========================================")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{Database: "sports_center", Username: "default", Password: ""},
	})
	if err != nil {
		log.Fatal("❌ Connect:", err)
	}
	defer conn.Close()
	fmt.Println("✅ Connected to ClickHouse")

	// Создаём схему
	for _, q := range []string{createTable, createMetrics, createMV} {
		if err := conn.Exec(ctx, q); err != nil {
			log.Fatal("❌ Schema:", err)
		}
	}
	fmt.Println("✅ Schema created")

	// Генерация данных
	var cnt uint64
	conn.QueryRow(ctx, "SELECT count() FROM sports_center.analytics_events").Scan(&cnt)
	if cnt < uint64(totalRecords) {
		fmt.Printf("  🔄 Generating %d events...\n", totalRecords)
		base := time.Now().AddDate(0, 0, -30)
		batch, _ := conn.PrepareBatch(ctx, `INSERT INTO sports_center.analytics_events (
			event_id,event_type,entity_type,entity_id,event_timestamp,client_id,coach_id,room_id,sport_type,
			price,refund_amount,currency,scheduled_at,cancelled_at,hours_before,source,promo_code,payment_method,
			cancel_reason,version,is_deleted)`)
		for i := 0; i < totalRecords; i++ {
			e := genEvent(base)
			batch.Append(e.EventID, e.EventType, e.EntityType, e.EntityID, e.EventTimestamp, e.ClientID,
				e.CoachID, e.RoomID, e.SportType, e.Price, e.RefundAmount, e.Currency, e.ScheduledAt,
				e.CancelledAt, e.HoursBefore, e.Source, e.PromoCode, e.PaymentMethod, e.CancelReason,
				e.Version, e.IsDeleted)
			if (i+1)%10000 == 0 {
				batch.Send()
				batch, _ = conn.PrepareBatch(ctx, `INSERT INTO sports_center.analytics_events (...)`)
				fmt.Printf("    ✓ %d / %d\n", i+1, totalRecords)
			}
		}
		batch.Send()
		time.Sleep(2 * time.Second)
	}
	fmt.Println("✅ Data loaded")

	// Выполнение запросов
	fmt.Println("\n🔍 Running analytical queries...")
	for _, q := range queries {
		rows, err := conn.Query(ctx, q.SQL)
		if err != nil {
			fmt.Printf("  ❌ %s: %v\n", q.Name, err)
			continue
		}
		cols := rows.ColumnTypes()
		fmt.Printf("  📊 %s\n", q.Desc)
		if len(cols) > 0 {
			headers := make([]string, len(cols))
			for i, c := range cols {
				headers[i] = c.Name()
			}
			fmt.Printf("     ┌─%s─┐\n", repeat("─┬─", len(headers)-1)+"─")
			limit := 0
			for rows.Next() && limit < 5 {
				vals := make([]interface{}, len(cols))
				ptrs := make([]interface{}, len(cols))
				for i := range cols {
					tn := cols[i].DatabaseTypeName()
					switch {
					case tn == "UInt8" || tn == "UInt16" || tn == "UInt32" || tn == "UInt64":
						var v uint64
						ptrs[i] = &v
					case tn == "Int8" || tn == "Int16" || tn == "Int32" || tn == "Int64":
						var v int64
						ptrs[i] = &v
					case tn == "Float32" || tn == "Float64":
						var v float64
						ptrs[i] = &v
					case tn == "String" || tn == "LowCardinality(String)" || tn == "Nullable(String)":
						var v string
						ptrs[i] = &v
					case tn == "Date" || tn == "DateTime" || tn == "DateTime('UTC')":
						var v time.Time
						ptrs[i] = &v
					default:
						var v interface{}
						ptrs[i] = &v
					}
				}
				rows.Scan(ptrs...)
				cells := make([]string, len(vals))
				for i, p := range ptrs {
					switch v := p.(type) {
					case *uint64:
						cells[i] = fmt.Sprintf("%v", *v)
					case *int64:
						cells[i] = fmt.Sprintf("%v", *v)
					case *float64:
						cells[i] = fmt.Sprintf("%.2f", *v)
					case *string:
						cells[i] = *v
					case *time.Time:
						cells[i] = (*v).Format("2006-01-02")
					default:
						cells[i] = fmt.Sprintf("%v", p)
					}
				}
				fmt.Printf("     │ %s │\n", join(cells, " │ "))
				limit++
			}
			fmt.Printf("     └─%s─┘\n", repeat("─┴─", len(headers)-1)+"─")
		}
		rows.Close()
	}

	// Сравнение производительности
	fmt.Println("\n⚡ Performance comparison:")
	for _, label := range []struct{ name, sql string }{
		{"Raw", `SELECT client_id, sum(price) as total FROM sports_center.analytics_events WHERE event_date>=today()-7 AND event_type IN ('TrainingBooked','SubscriptionPurchased') GROUP BY client_id ORDER BY total DESC LIMIT 10`},
		{"View", `SELECT client_id, sum(total_spent) as total FROM sports_center.daily_client_metrics WHERE metric_date>=today()-7 GROUP BY client_id ORDER BY total DESC LIMIT 10`},
	} {
		start := time.Now()
		rows, _ := conn.Query(ctx, label.sql)
		n := 0
		for rows.Next() {
			n++
		}
		rows.Close()
		dur := time.Since(start)
		fmt.Printf("  %s: %d rows, %v\n", label.name, n, dur.Round(time.Millisecond))
	}

	// Ключевые метрики
	fmt.Println("\n📈 Key Metrics:")
	metrics := []struct{ name, sql string }{
		{"Revenue", `SELECT formatDecimal(round(sumIf(price, event_type IN ('TrainingBooked','SubscriptionPurchased')),2),2) || ' RUB' FROM sports_center.analytics_events WHERE event_date>=today()-30`},
		{"Clients", `SELECT toString(uniqIf(client_id, event_date>=today()-30)) FROM sports_center.analytics_events`},
		{"Avg Price", `SELECT formatDecimal(round(avgIf(price, event_type='TrainingBooked'),2),2) || ' RUB' FROM sports_center.analytics_events WHERE event_date>=today()-30`},
		{"Cancel Rate", `SELECT concat(toString(round(countIf(event_type='TrainingCancelled')*100/nullIf(countIf(event_type IN ('TrainingBooked','TrainingCancelled')),0),2)),'%') FROM sports_center.analytics_events WHERE event_date>=today()-30`},
		{"Top Sport", `SELECT ifNull(argMax(sport_type,cnt),'N/A') FROM (SELECT sport_type,count() as cnt FROM sports_center.analytics_events WHERE event_type='TrainingBooked' AND event_date>=today()-30 GROUP BY sport_type)`},
	}
	for _, m := range metrics {
		var v string
		conn.QueryRow(ctx, m.sql).Scan(&v)
		if v == "" {
			v = "N/A"
		}
		fmt.Printf("  • %s: %s\n", m.name, v)
	}

	fmt.Println("\n🎉 All done!")
}

func repeat(s string, n int) string {
	r := ""
	for i := 0; i < n; i++ {
		r += s
	}
	return r
}

func join(a []string, sep string) string {
	if len(a) == 0 {
		return ""
	}
	r := a[0]
	for _, v := range a[1:] {
		r += sep + v
	}
	return r
}
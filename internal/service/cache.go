package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func getOrSetJSON[T any](ctx context.Context, rdb redis.UniversalClient, key string, ttl time.Duration, fn func() (*T, error)) (*T, error) {
	cacheVal, err := rdb.Get(ctx, key).Result()
	if err == nil {
		var result T
		if err := json.Unmarshal([]byte(cacheVal), &result); err != nil {
			log.Printf("Redis unmarshal error for %s: %v", key, err)
		} else {
			fmt.Println("REDIS HIT:", key)
			return &result, nil
		}
	}

	fmt.Println("REDIS MISS:", key)
	result, err := fn()
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(result); err == nil {
		rdb.Set(ctx, key, string(data), ttl)
	}
	return result, nil
}

func GetTotalRevenue(db *sql.DB, rdb redis.UniversalClient, ctx context.Context) (float64, error) {
	val, err := getOrSetJSON(ctx, rdb, "report:total_revenue", 10*time.Second, func() (*float64, error) {
		var t float64
		err := db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount),0) FROM payments WHERE status='completed'").Scan(&t)
		if err != nil {
			return nil, err
		}
		return &t, nil
	})
	if err != nil {
		return 0, err
	}
	return *val, nil
}

func GetAvgClassRating(db *sql.DB, rdb redis.UniversalClient, ctx context.Context) (float64, error) {
	val, err := getOrSetJSON(ctx, rdb, "report:avg_rating", 10*time.Second, func() (*float64, error) {
		var t float64
		err := db.QueryRowContext(ctx, "SELECT COALESCE(AVG(rating),0) FROM reviews").Scan(&t)
		if err != nil {
			return nil, err
		}
		return &t, nil
	})
	if err != nil {
		return 0, err
	}
	return *val, nil
}

type TopSport struct {
	Name   string `json:"name"`
	Visits int    `json:"visits"`
}

func GetTopSportsByAttendance(db *sql.DB, rdb redis.UniversalClient, ctx context.Context) ([]TopSport, error) {
	result, err := getOrSetJSON(ctx, rdb, "report:top_sports", 2*time.Minute, func() (*[]TopSport, error) {
		const query = `
			SELECT sp.name, COUNT(*) AS visits
			FROM attendance_logs al
			JOIN bookings b ON al.user_id = b.user_id
			JOIN schedules s ON b.schedule_id = s.id
			JOIN classes c ON s.class_id = c.id
			JOIN sports sp ON c.sport_id = sp.id
			GROUP BY sp.name
			ORDER BY visits DESC
			LIMIT 5
		`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []TopSport
		for rows.Next() {
			var name string
			var visits int
			if err := rows.Scan(&name, &visits); err != nil {
				return nil, err
			}
			items = append(items, TopSport{Name: name, Visits: visits})
		}
		return &items, nil
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}

type UserRank struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
	Rank   int `json:"rank"`
}

func GetUserRankByLoyalty(db *sql.DB, rdb redis.UniversalClient, ctx context.Context) ([]UserRank, error) {
	result, err := getOrSetJSON(ctx, rdb, "report:user_rank_loyalty", 60*time.Second, func() (*[]UserRank, error) {
		const query = `
			SELECT user_id, points,
			RANK() OVER (ORDER BY points DESC) AS rank
			FROM loyalty_points
			ORDER BY rank
			LIMIT 10
		`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []UserRank
		for rows.Next() {
			var uid, pts, rank int
			if err := rows.Scan(&uid, &pts, &rank); err != nil {
				return nil, err
			}
			items = append(items, UserRank{UserID: uid, Points: pts, Rank: rank})
		}
		return &items, nil
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}

type CoachRating struct {
	CoachID   int     `json:"coach_id"`
	AvgRating float64 `json:"avg_rating"`
	Rank      int     `json:"rank"`
}

func GetCoachRatingWithRowNumber(db *sql.DB, rdb redis.UniversalClient, ctx context.Context) ([]CoachRating, error) {
	result, err := getOrSetJSON(ctx, rdb, "report:coach_rating", 90*time.Second, func() (*[]CoachRating, error) {
		const query = `
			SELECT coach_id, AVG(rating) AS avg_rating,
			ROW_NUMBER() OVER (ORDER BY AVG(rating) DESC) AS rn
			FROM reviews
			WHERE coach_id IS NOT NULL
			GROUP BY coach_id
			HAVING AVG(rating) >= 3.0
			ORDER BY avg_rating DESC
			LIMIT 5
		`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []CoachRating
		for rows.Next() {
			var cid int
			var avg float64
			var rn int
			if err := rows.Scan(&cid, &avg, &rn); err != nil {
				return nil, err
			}
			items = append(items, CoachRating{CoachID: cid, AvgRating: avg, Rank: rn})
		}
		return &items, nil
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}

type BookingsPerDay struct {
	Day      string `json:"day"`
	Bookings int    `json:"bookings"`
}

func GetBookingsPerDay(db *sql.DB, rdb redis.UniversalClient, ctx context.Context) ([]BookingsPerDay, error) {
	result, err := getOrSetJSON(ctx, rdb, "report:bookings_per_day", 60*time.Second, func() (*[]BookingsPerDay, error) {
		const query = `
			SELECT DATE(start_time) AS day, COUNT(*) AS bookings
			FROM schedules s
			JOIN bookings b ON s.id = b.schedule_id
			GROUP BY day
			ORDER BY day DESC
			LIMIT 7
		`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []BookingsPerDay
		for rows.Next() {
			var day time.Time
			var cnt int
			if err := rows.Scan(&day, &cnt); err != nil {
				return nil, err
			}
			items = append(items, BookingsPerDay{
				Day:      day.Format("2006-01-02"),
				Bookings: cnt,
			})
		}
		return &items, nil
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}

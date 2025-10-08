package service

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// =============== БИЗНЕС-ЗАПРОСЫ ===============

// --- Агрегирующие (4) ---
func GetTotalRevenue(db *sql.DB) float64 {
	var total float64
	db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed'
	`).Scan(&total)
	return total
}

func GetAvgClassRating(db *sql.DB) float64 {
	var avg float64
	db.QueryRow("SELECT COALESCE(AVG(rating), 0) FROM reviews").Scan(&avg)
	return avg
}

func GetBookingsPerDay(db *sql.DB) {
	const query = `
		SELECT DATE(start_time) AS day, COUNT(*) AS bookings
		FROM schedules s
		JOIN bookings b ON s.id = b.schedule_id
		GROUP BY day
		ORDER BY day DESC
		LIMIT 7
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var day time.Time
		var cnt int
		rows.Scan(&day, &cnt)
		fmt.Printf("📅 %s: %d bookings\n", day.Format("2006-01-02"), cnt)
	}
}

func GetTopSportsByAttendance(db *sql.DB) {
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

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var name string
		var visits int
		rows.Scan(&name, &visits)
		fmt.Printf("🏆 %s: %d visits\n", name, visits)
	}
}

// --- Оконные функции (4) ---
func GetUserRankByLoyalty(db *sql.DB) {
	const query = `
		SELECT user_id, points,
		RANK() OVER (ORDER BY points DESC) AS rank
		FROM loyalty_points
		ORDER BY rank
		LIMIT 10
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var uid, pts, rank int
		rows.Scan(&uid, &pts, &rank)
		fmt.Printf("🏅 User %d: %d pts (rank %d)\n", uid, pts, rank)
	}
}

func GetRunningTotalRevenue(db *sql.DB) {
	const query = `
		SELECT id, amount, 
		SUM(amount) OVER (ORDER BY id) AS running_total
		FROM payments
		WHERE status = 'completed'
		ORDER BY id
		LIMIT 5
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var id int
		var amt, total float64
		rows.Scan(&id, &amt, &total)
		fmt.Printf("💰 Payment %d: $%.2f → Total: $%.2f\n", id, amt, total)
	}
}

func GetClassBookingsWithMovingAvg(db *sql.DB) {
	const query = `
		SELECT c.id, COUNT(b.id) AS bookings,
		AVG(COUNT(b.id)) OVER 
		(ORDER BY c.id ROWS BETWEEN 2 PRECEDING AND CURRENT ROW) 
		AS moving_avg
		FROM classes c
		LEFT JOIN schedules s ON c.id = s.class_id
		LEFT JOIN bookings b ON s.id = b.schedule_id
		GROUP BY c.id
		ORDER BY c.id
		LIMIT 10
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var cid, bookings int
		var avg float64
		rows.Scan(&cid, &bookings, &avg)
		fmt.Printf("📚 Class %d: %d bookings (avg: %.2f)\n", cid, bookings, avg)
	}
}

func GetCoachRatingWithRowNumber(db *sql.DB) {
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

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var cid int
		var avg float64
		var rn int
		rows.Scan(&cid, &avg, &rn)
		fmt.Printf("👨‍🏫 Coach %d: %.2f ★ (rank %d)\n", cid, avg, rn)
	}
}

// --- JOIN 2 таблицы (2) ---
func GetUsersWithLoyalty(db *sql.DB) {
	const query = `
		SELECT u.email, lp.points
		FROM users u
		JOIN loyalty_points lp ON u.id = lp.user_id
		LIMIT 5
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var email string
		var pts int
		rows.Scan(&email, &pts)
		fmt.Printf("📧 %s → %d pts\n", email, pts)
	}
}

func GetActiveMemberships(db *sql.DB) {
	const query = `
		SELECT u.email, m.duration_days, um.started_at
		FROM users u
		JOIN user_memberships um ON u.id = um.user_id
		JOIN memberships m ON um.membership_id = m.id
		WHERE um.is_active = true
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("GetActiveMemberships query error:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		var days int
		var started time.Time
		if err := rows.Scan(&email, &days, &started); err != nil {
			log.Println("GetActiveMemberships scan error:", err)
			continue
		}
		fmt.Printf("🎫 %s: %d days since %s\n", email, days, started.Format("2006-01-02"))
	}
}

// --- JOIN 3 таблицы (4) ---
func GetBookingsWithDetails(db *sql.DB) {
	const query = `
		SELECT u.email, sp.name AS sport, s.start_time
		FROM bookings b
		JOIN users u ON b.user_id = u.id
		JOIN schedules s ON b.schedule_id = s.id
		JOIN classes c ON s.class_id = c.id
		JOIN sports sp ON c.sport_id = sp.id
		LIMIT 5
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var email, sport string
		var start time.Time
		rows.Scan(&email, &sport, &start)
		fmt.Printf("📅 %s booked %s at %s\n", email, sport, start.Format("15:04"))
	}
}

func GetPaymentsWithMembership(db *sql.DB) {
	const query = `
		SELECT u.email, p.amount, m.duration_days
		FROM payments p
		JOIN users u ON p.user_id = u.id
		JOIN user_memberships um ON p.user_id = um.user_id
		JOIN memberships m ON um.membership_id = m.id
		LIMIT 5
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var email string
		var amt float64
		var days int
		rows.Scan(&email, &amt, &days)
		fmt.Printf("💳 %s paid $%.2f for %d-day plan\n", email, amt, days)
	}
}

func GetReviewsWithCoachInfo(db *sql.DB) {
	const query = `
		SELECT u.email, r.rating, 'Coach ' || r.coach_id AS coach
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.coach_id IS NOT NULL
		LIMIT 5
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var email string
		var rating int
		var coach string
		rows.Scan(&email, &rating, &coach)
		fmt.Printf("⭐ %s rated %s: %d\n", email, coach, rating)
	}
}

func GetReferralRewards(db *sql.DB) {
	const query = `
		SELECT ref.email AS referrer, refd.email AS referred, r.rewarded
		FROM referrals r
		JOIN users ref ON r.referrer_id = ref.id
		JOIN users refd ON r.referred_id = refd.id
		LIMIT 5
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var ref, refd string
		var rewarded bool
		rows.Scan(&ref, &refd, &rewarded)
		fmt.Printf("🤝 %s referred %s (rewarded: %t)\n", ref, refd, rewarded)
	}
}

// --- JOIN 4 таблицы (1) ---
func GetScheduleWithRoomAndSport(db *sql.DB) {
	const query = `
		SELECT s.start_time, sp.name AS sport, r.capacity, c.coach_id
		FROM schedules s
		JOIN classes c ON s.class_id = c.id
		JOIN sports sp ON c.sport_id = sp.id
		JOIN rooms r ON s.room_id = r.id
		LIMIT 5
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var start time.Time
		var sport string
		var cap int
		var coach int
		rows.Scan(&start, &sport, &cap, &coach)
		fmt.Printf("🏋️ %s at %s in room (cap %d) by coach %d\n",
		          sport, start.Format("15:04"), cap, coach)
	}
}

// --- JOIN 5 таблиц (1) ---
func GetFullBookingInfo(db *sql.DB) {
	const query = `
		SELECT u.email, sp.name AS sport, c.coach_id, r.capacity, s.start_time
		FROM bookings b
		JOIN users u ON b.user_id = u.id
		JOIN schedules s ON b.schedule_id = s.id
		JOIN classes c ON s.class_id = c.id
		JOIN sports sp ON c.sport_id = sp.id
		JOIN rooms r ON s.room_id = r.id
		LIMIT 5
	`

	rows, _ := db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var email, sport string
		var coach, cap int
		var start time.Time
		rows.Scan(&email, &sport, &coach, &cap, &start)
		fmt.Printf("✅ %s booked %s (coach %d) in room (cap %d) at %s\n",
		          email, sport, coach, cap, start.Format("15:04"))
	}
}

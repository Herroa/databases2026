package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"os"
	"flag"

	_ "github.com/lib/pq"
)

type dBConnectionSettings struct {
	host         string
	port         int
	user         string
	password     string
	dataBaseName string
}

var commonDb = dBConnectionSettings{
	host:         "localhost",
	port:         5432,
	user:         "postgres",
	password:     "postgres",
	dataBaseName: "postgres",
}

var sportsDb = dBConnectionSettings{
	host:         "localhost",
	port:         5432,
	user:         "postgres",
	password:     "postgres",
	dataBaseName: "sports_club",
}

var dataBase *sql.DB

func initDataBase(dbSettings dBConnectionSettings) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	    dbSettings.host, dbSettings.port, dbSettings.user, dbSettings.password, dbSettings.dataBaseName)
	var err error
	dataBase, err = sql.Open("postgres", psqlInfo)
	fmt.Println(psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	if err = dataBase.Ping(); err != nil {
		log.Fatal(err)
	}
}

// =============== CRUD ДЛЯ ВСЕХ ТАБЛИЦ ===============

// --- 1. users ---
func CreateUser(email string) (int, error) {
	var id int
	err := dataBase.QueryRow("INSERT INTO users (email) VALUES ($1) RETURNING id", email).Scan(&id)
	return id, err
}
func DeleteUser(id int) error {
	_, err := dataBase.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// --- 2. coaches ---
func CreateCoach(userID int) error {
	_, err := dataBase.Exec("INSERT INTO coaches (user_id) VALUES ($1)", userID)
	return err
}
func DeleteCoach(userID int) error {
	_, err := dataBase.Exec("DELETE FROM coaches WHERE user_id = $1", userID)
	return err
}

// --- 3. sports ---
func CreateSport(name string) (int, error) {
	var id int
	err := dataBase.QueryRow("INSERT INTO sports (name) VALUES ($1) RETURNING id",
	                        name).Scan(&id)
	return id, err
}
func DeleteSport(id int) error {
	_, err := dataBase.Exec("DELETE FROM sports WHERE id = $1", id)
	return err
}

// --- 4. classes ---
func CreateClass(sportID, coachID int) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO classes (sport_id, coach_id) 
	                          VALUES ($1, $2) RETURNING id`,
	                        sportID, coachID).Scan(&id)
	return id, err
}
func DeleteClass(id int) error {
	_, err := dataBase.Exec("DELETE FROM classes WHERE id = $1", id)
	return err
}

// --- 5. rooms ---
func CreateRoom(capacity int) (int, error) {
	var id int
	err := dataBase.QueryRow("INSERT INTO rooms (capacity) VALUES ($1) RETURNING id",
	                        capacity).Scan(&id)
	return id, err
}
func DeleteRoom(id int) error {
	_, err := dataBase.Exec("DELETE FROM rooms WHERE id = $1", id)
	return err
}

// --- 6. schedules ---
func CreateSchedule(classID, roomID int, start, end time.Time) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO schedules (class_id, room_id, start_time, end_time)
	                          VALUES ($1, $2, $3, $4) RETURNING id`,
	                        classID, roomID, start, end).Scan(&id)
	return id, err
}
func DeleteSchedule(id int) error {
	_, err := dataBase.Exec("DELETE FROM schedules WHERE id = $1", id)
	return err
}

// --- 7. bookings ---
func CreateBooking(userID, scheduleID int) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO bookings (user_id, schedule_id)
	                          VALUES ($1, $2) RETURNING id`,
	                        userID, scheduleID).Scan(&id)
	return id, err
}
func DeleteBooking(id int) error {
	_, err := dataBase.Exec("DELETE FROM bookings WHERE id = $1", id)
	return err
}

// --- 8. memberships ---
func CreateMembership(durationDays int, price float64) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO memberships (duration_days, price) 
	                          VALUES ($1, $2) RETURNING id`,
	                        durationDays, price).Scan(&id)
	return id, err
}
func DeleteMembership(id int) error {
	_, err := dataBase.Exec("DELETE FROM memberships WHERE id = $1", id)
	return err
}

// --- 9. user_memberships ---
func CreateUserMembership(userID, membershipID int, started, ended time.Time) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO user_memberships (user_id, membership_id, started_at, ended_at)
	                          VALUES ($1, $2, $3, $4) RETURNING id`,
	                        userID, membershipID, started, ended).Scan(&id)
	return id, err
}
func DeleteUserMembership(id int) error {
	_, err := dataBase.Exec("DELETE FROM user_memberships WHERE id = $1", id)
	return err
}

// --- 10. payments ---
func CreatePayment(userID int, amount float64) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO payments (user_id, amount)
	                          VALUES ($1, $2) RETURNING id`,
	                        userID, amount).Scan(&id)
	return id, err
}
func DeletePayment(id int) error {
	_, err := dataBase.Exec("DELETE FROM payments WHERE id = $1", id)
	return err
}

// --- 11. attendance_logs ---
func CreateAttendanceLog(userID int, start, end time.Time) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO attendance_logs (user_id, start_time, end_time)
	                          VALUES ($1, $2, $3) RETURNING id`,
	                        userID, start, end).Scan(&id)
	return id, err
}
func DeleteAttendanceLog(id int) error {
	_, err := dataBase.Exec("DELETE FROM attendance_logs WHERE id = $1", id)
	return err
}

// --- 12. reviews ---
func CreateReview(userID, coachID, classID int, rating int) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO reviews (user_id, coach_id, class_id, rating)
	                          VALUES ($1, $2, $3, $4) RETURNING id`,
	                        userID, nullInt(coachID), nullInt(classID), rating).Scan(&id)
	return id, err
}
func DeleteReview(id int) error {
	_, err := dataBase.Exec("DELETE FROM reviews WHERE id = $1", id)
	return err
}

// --- 13. promotions ---
func CreatePromotion(code string, discount int, from, until time.Time, maxUses *int) (int, error) {
	var id int
	if maxUses == nil {
		err := dataBase.QueryRow(`INSERT INTO promotions 
		                          (code, discount_percent, valid_from, valid_until)
		                          VALUES ($1, $2, $3, $4) RETURNING id`,
		                        code, discount, from, until).Scan(&id)
		return id, err
	}
	err := dataBase.QueryRow(`INSERT INTO promotions 
	                          (code, discount_percent, valid_from, valid_until, max_uses)
	                          VALUES ($1, $2, $3, $4, $5) RETURNING id`,
	                        code, discount, from, until, *maxUses).Scan(&id)
	return id, err
}
func DeletePromotion(id int) error {
	_, err := dataBase.Exec("DELETE FROM promotions WHERE id = $1", id)
	return err
}

// --- 14. promotion_usage ---
func UsePromotion(userID, promoID int) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO promotion_usage (user_id, promotion_id) 
	                          VALUES ($1, $2) RETURNING id`,
	                        userID, promoID).Scan(&id)
	return id, err
}
func DeletePromotionUsage(id int) error {
	_, err := dataBase.Exec("DELETE FROM promotion_usage WHERE id = $1", id)
	return err
}

// --- 15. notifications ---
func CreateNotification(userID int, isRead bool) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO notifications (user_id, is_read) 
	                          VALUES ($1, $2) RETURNING id`,
	                        userID, isRead).Scan(&id)
	return id, err
}
func DeleteNotification(id int) error {
	_, err := dataBase.Exec("DELETE FROM notifications WHERE id = $1", id)
	return err
}

// --- 16. loyalty_points ---
func SetLoyaltyPoints(userID, points int) error {
	_, err := dataBase.Exec(`INSERT INTO loyalty_points (user_id, points)
	                         VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET points = $2`,
	                       userID, points)
	return err
}
func DeleteLoyaltyPoints(userID int) error {
	_, err := dataBase.Exec("DELETE FROM loyalty_points WHERE user_id = $1", userID)
	return err
}

// --- 17. referrals ---
func CreateReferral(referrerID, referredID int) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO referrals (referrer_id, referred_id) 
	                          VALUES ($1, $2) RETURNING id`,
	                        referrerID, referredID).Scan(&id)
	return id, err
}
func DeleteReferral(id int) error {
	_, err := dataBase.Exec("DELETE FROM referrals WHERE id = $1", id)
	return err
}

// --- 18. audit_logs ---
func LogAudit(userID *int, action, entityType string, entityID *int) (int, error) {
	var id int
	if userID == nil && entityID == nil {
		err := dataBase.QueryRow("INSERT INTO audit_logs (action) VALUES ($1) RETURNING id",
		                        action).Scan(&id)
		return id, err
	}
	if userID == nil {
		err := dataBase.QueryRow(`INSERT INTO audit_logs (action, entity_type, entity_id) 
		                          VALUES ($1, $2, $3) RETURNING id`,
		                        action, entityType, entityID).Scan(&id)
		return id, err
	}
	if entityID == nil {
		err := dataBase.QueryRow(`INSERT INTO audit_logs (user_id, action, entity_type) 
		                          VALUES ($1, $2, $3) RETURNING id`,
		                        *userID, action, entityType).Scan(&id)
		return id, err
	}
	err := dataBase.QueryRow(`INSERT INTO audit_logs (user_id, action, entity_type, entity_id) 
	                          VALUES ($1, $2, $3, $4) RETURNING id`,
	                        *userID, action, entityType, *entityID).Scan(&id)
	return id, err
}
func DeleteAuditLog(id int) error {
	_, err := dataBase.Exec("DELETE FROM audit_logs WHERE id = $1", id)
	return err
}

// --- 19. system_settings ---
func SetSystemSetting(key, value string) error {
	_, err := dataBase.Exec(`INSERT INTO system_settings (key, value)
	                         VALUES ($1, $2)
	                         ON CONFLICT (key) DO UPDATE SET value = $`,
	                       key, value)
	return err
}
func DeleteSystemSetting(key string) error {
	_, err := dataBase.Exec("DELETE FROM system_settings WHERE key = $1", key)
	return err
}

// --- 20. temp_bookings ---
func CreateTempBooking(userID, scheduleID int, expires time.Time, token string) (int, error) {
	var id int
	err := dataBase.QueryRow(`INSERT INTO temp_bookings (user_id, schedule_id, expires_at, token)
	                          VALUES ($1, $2, $3, $4) RETURNING id`,
	                        userID, scheduleID, expires, token).Scan(&id)
	return id, err
}
func DeleteTempBooking(id int) error {
	_, err := dataBase.Exec("DELETE FROM temp_bookings WHERE id = $1", id)
	return err
}

// =============== ВСПОМОГАТЕЛЬНЫЕ ===============
func nullInt(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

// =============== БИЗНЕС-ЗАПРОСЫ ===============

// --- Агрегирующие (4) ---
func GetTotalRevenue() float64 {
	var total float64
	dataBase.QueryRow(`SELECT COALESCE(SUM(amount), 0) 
	                   FROM payments WHERE status = 'completed'`).Scan(&total)
	return total
}

func GetAvgClassRating() float64 {
	var avg float64
	dataBase.QueryRow("SELECT COALESCE(AVG(rating), 0) FROM reviews").Scan(&avg)
	return avg
}

func GetBookingsPerDay() {
	rows, _ := dataBase.Query(`SELECT DATE(start_time) AS day, COUNT(*) AS bookings
	                           FROM schedules s
	                           JOIN bookings b ON s.id = b.schedule_id
	                           GROUP BY day
	                           ORDER BY day DESC
	                           LIMIT 7`)
	defer rows.Close()
	for rows.Next() {
		var day time.Time
		var cnt int
		rows.Scan(&day, &cnt)
		fmt.Printf("📅 %s: %d bookings\n", day.Format("2006-01-02"), cnt)
	}
}

func GetTopSportsByAttendance() {
	rows, _ := dataBase.Query(`SELECT sp.name, COUNT(*) AS visits
	                           FROM attendance_logs al
	                           JOIN bookings b ON al.user_id = b.user_id
	                           JOIN schedules s ON b.schedule_id = s.id
	                           JOIN classes c ON s.class_id = c.id
	                           JOIN sports sp ON c.sport_id = sp.id
	                           GROUP BY sp.name
	                           ORDER BY visits DESC
	                           LIMIT 5`)
	defer rows.Close()
	for rows.Next() {
		var name string
		var visits int
		rows.Scan(&name, &visits)
		fmt.Printf("🏆 %s: %d visits\n", name, visits)
	}
}

// --- Оконные функции (4) ---
func GetUserRankByLoyalty() {
	rows, _ := dataBase.Query(`SELECT user_id, points,
	                           RANK() OVER (ORDER BY points DESC) AS rank
	                           FROM loyalty_points
	                           ORDER BY rank
	                           LIMIT 10`)
	defer rows.Close()
	for rows.Next() {
		var uid, pts, rank int
		rows.Scan(&uid, &pts, &rank)
		fmt.Printf("🏅 User %d: %d pts (rank %d)\n", uid, pts, rank)
	}
}

func GetRunningTotalRevenue() {
	rows, _ := dataBase.Query(`SELECT id, amount, 
	                           SUM(amount) OVER (ORDER BY id) AS running_total
	                           FROM payments
	                           WHERE status = 'completed'
	                           ORDER BY id
	                           LIMIT 5`)
	defer rows.Close()
	for rows.Next() {
		var id int
		var amt, total float64
		rows.Scan(&id, &amt, &total)
		fmt.Printf("💰 Payment %d: $%.2f → Total: $%.2f\n", id, amt, total)
	}
}

func GetClassBookingsWithMovingAvg() {
	rows, _ := dataBase.Query(`SELECT c.id, COUNT(b.id) AS bookings,
	                           AVG(COUNT(b.id)) OVER 
	                           (ORDER BY c.id ROWS BETWEEN 2 PRECEDING AND CURRENT ROW) 
	                           AS moving_avg
	                           FROM classes c
	                           LEFT JOIN schedules s ON c.id = s.class_id
	                           LEFT JOIN bookings b ON s.id = b.schedule_id
	                           GROUP BY c.id
	                           ORDER BY c.id
	                           LIMIT 10`)
	defer rows.Close()
	for rows.Next() {
		var cid, bookings int
		var avg float64
		rows.Scan(&cid, &bookings, &avg)
		fmt.Printf("📚 Class %d: %d bookings (avg: %.2f)\n", cid, bookings, avg)
	}
}

func GetCoachRatingWithRowNumber() {
	rows, _ := dataBase.Query(`SELECT coach_id, AVG(rating) AS avg_rating,
	                           ROW_NUMBER() OVER (ORDER BY AVG(rating) DESC) AS rn
	                           FROM reviews
	                           WHERE coach_id IS NOT NULL
	                           GROUP BY coach_id
	                           HAVING AVG(rating) >= 3.0
	                           ORDER BY avg_rating DESC
	                           LIMIT 5`)
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
func GetUsersWithLoyalty() {
	rows, _ := dataBase.Query(`SELECT u.email, lp.points
	                           FROM users u
	                           JOIN loyalty_points lp ON u.id = lp.user_id
	                           LIMIT 5`)
	defer rows.Close()
	for rows.Next() {
		var email string
		var pts int
		rows.Scan(&email, &pts)
		fmt.Printf("📧 %s → %d pts\n", email, pts)
	}
}

func GetActiveMemberships() {
	rows, err := dataBase.Query(`SELECT u.email, m.duration_days, um.started_at
	                             FROM users u
	                             JOIN user_memberships um ON u.id = um.user_id
	                             JOIN memberships m ON um.membership_id = m.id
	                             WHERE um.is_active = true
	                             LIMIT 5`)
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
func GetBookingsWithDetails() {
	rows, _ := dataBase.Query(`SELECT u.email, sp.name AS sport, s.start_time
	                           FROM bookings b
	                           JOIN users u ON b.user_id = u.id
	                           JOIN schedules s ON b.schedule_id = s.id
	                           JOIN classes c ON s.class_id = c.id
	                           JOIN sports sp ON c.sport_id = sp.id
	                           LIMIT 5`)
	defer rows.Close()
	for rows.Next() {
		var email, sport string
		var start time.Time
		rows.Scan(&email, &sport, &start)
		fmt.Printf("📅 %s booked %s at %s\n", email, sport, start.Format("15:04"))
	}
}

func GetPaymentsWithMembership() {
	rows, _ := dataBase.Query(`SELECT u.email, p.amount, m.duration_days
	                           FROM payments p
	                           JOIN users u ON p.user_id = u.id
	                           JOIN user_memberships um ON p.user_id = um.user_id
	                           JOIN memberships m ON um.membership_id = m.id
	                           LIMIT 5`)
	defer rows.Close()
	for rows.Next() {
		var email string
		var amt float64
		var days int
		rows.Scan(&email, &amt, &days)
		fmt.Printf("💳 %s paid $%.2f for %d-day plan\n", email, amt, days)
	}
}

func GetReviewsWithCoachInfo() {
	rows, _ := dataBase.Query(`SELECT u.email, r.rating, 'Coach ' || r.coach_id AS coach
	                           FROM reviews r
	                           JOIN users u ON r.user_id = u.id
	                           WHERE r.coach_id IS NOT NULL
	                           LIMIT 5`)
	defer rows.Close()
	for rows.Next() {
		var email string
		var rating int
		var coach string
		rows.Scan(&email, &rating, &coach)
		fmt.Printf("⭐ %s rated %s: %d\n", email, coach, rating)
	}
}

func GetReferralRewards() {
	rows, _ := dataBase.Query(`SELECT ref.email AS referrer, refd.email AS referred, r.rewarded
	                           FROM referrals r
	                           JOIN users ref ON r.referrer_id = ref.id
	                           JOIN users refd ON r.referred_id = refd.id
	                           LIMIT 5`)
	defer rows.Close()
	for rows.Next() {
		var ref, refd string
		var rewarded bool
		rows.Scan(&ref, &refd, &rewarded)
		fmt.Printf("🤝 %s referred %s (rewarded: %t)\n", ref, refd, rewarded)
	}
}

// --- JOIN 4 таблицы (1) ---
func GetScheduleWithRoomAndSport() {
	rows, _ := dataBase.Query(`SELECT s.start_time, sp.name AS sport, r.capacity, c.coach_id
	                           FROM schedules s
	                           JOIN classes c ON s.class_id = c.id
	                           JOIN sports sp ON c.sport_id = sp.id
	                           JOIN rooms r ON s.room_id = r.id
	                           LIMIT 5`)
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
func GetFullBookingInfo() {
	rows, _ := dataBase.Query(`SELECT u.email, sp.name AS sport, c.coach_id, r.capacity, s.start_time
	                           FROM bookings b
	                           JOIN users u ON b.user_id = u.id
	                           JOIN schedules s ON b.schedule_id = s.id
	                           JOIN classes c ON s.class_id = c.id
	                           JOIN sports sp ON c.sport_id = sp.id
	                           JOIN rooms r ON s.room_id = r.id
	                           LIMIT 5`)
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

func DbIsExist() bool {
	var exists bool
	err := dataBase.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname='sports_club');").Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check database existence: %v", err)
		os.Exit(1)
	}
	return exists
}

func testSportClubDb() {
	initDataBase(sportsDb)
	defer dataBase.Close()

	if !DbIsExist() {
		log.Fatalf("'sports_club' database doesn't exists")
		os.Exit(1)
	}

	fmt.Println("✅ Подключено к 'sports_club' БД")

	// crud
	userID, err := CreateUser("test78@example.com")
	if (err != nil) {
		fmt.Println("CreateUser: ", err)
		os.Exit(1)
	}

	err = CreateCoach(userID)
	if (err != nil) {
		fmt.Println("CreateCoach: ", err)
		os.Exit(1)
	}

	sportID, err := CreateSport("Pilates")
	if (err != nil) {
		fmt.Println("CreateSport: ", err)
		os.Exit(1)
	}

	classID, err := CreateClass(sportID, userID)
	if (err != nil) {
		fmt.Println("CreateClass: ", err)
		os.Exit(1)
	}

	roomID, err := CreateRoom(20)
	if (err != nil) {
		fmt.Println("CreateRoom: ", err)
		os.Exit(1)
	}

	schedID, err := CreateSchedule(classID, roomID, time.Now(), time.Now().Add(time.Hour))
	if (err != nil) {
		fmt.Println("CreateSchedule: ", err)
		os.Exit(1)
	}

	bookingID, err := CreateBooking(userID, schedID)
	if (err != nil) {
		fmt.Println("CreateBooking: ", err)
		os.Exit(1)
	}

	fmt.Printf("Созданы сущности: user=%d, booking=%d\n", userID, bookingID)

	// business
	fmt.Println("\n📊 Агрегирующие:")
	fmt.Printf("Общий доход: $%.2f\n", GetTotalRevenue())
	fmt.Printf("Средний рейтинг: %.2f\n", GetAvgClassRating())
	GetBookingsPerDay()
	GetTopSportsByAttendance()

	fmt.Println("\n🪟 Оконные функции:")
	GetUserRankByLoyalty()
	GetRunningTotalRevenue()
	GetClassBookingsWithMovingAvg()
	GetCoachRatingWithRowNumber()

	fmt.Println("\n🔗 JOIN-запросы:")
	GetUsersWithLoyalty()
	GetActiveMemberships()
	GetBookingsWithDetails()
	GetPaymentsWithMembership()
	GetReviewsWithCoachInfo()
	GetReferralRewards()
	GetScheduleWithRoomAndSport()
	GetFullBookingInfo()

	// Очистка
	DeleteBooking(bookingID)
	DeleteSchedule(schedID)
	DeleteClass(classID)
	DeleteSport(sportID)
	DeleteCoach(userID)
	DeleteUser(userID)
}

func initSportsDb() {
	initDataBase(commonDb)

	fmt.Println("✅ Подключено к 'postgres' БД")

	// Проверяем, существует ли база данных
	if DbIsExist() {
		log.Fatalf("Database sports_club already exists")
		dataBase.Close()
		os.Exit(1)
	}

	// Создаём базу данных
	_, err := dataBase.Exec(`CREATE DATABASE sports_club 
	                        ENCODING 'UTF8' 
	                        LC_COLLATE 'en_US.UTF-8' 
	                        LC_CTYPE 'en_US.UTF-8' 
	                        TEMPLATE template0;`)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
		dataBase.Close()
	}
	fmt.Println("Database sports_club created")
	dataBase.Close()

	// Подключаемся к базе sports_club
	initDataBase(sportsDb)
	fmt.Println("✅ Подключено к 'sports_club' БД")
	defer dataBase.Close()

	// Функция для выполнения SQL из файла
	execSQLFile := func(filename string) {
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filename, err)
			os.Exit(1)
		}
		_, err = dataBase.Exec(string(content))
		if err != nil {
			log.Fatalf("Failed to execute file %s: %v", filename, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Executed %s successfully\n", filename)
	}

	// Выполняем init_db.sql
	execSQLFile("../configs/sql/init_db.sql")
	// Выполняем generate_3m_bookings.sql
	execSQLFile("../configs/sql/generate_3m_bookings.sql")
}

func main() {
	initFlag := flag.Bool("init", false, "Initialization of 'sports_club' database")
	testFlag := flag.Bool("test", false, "Test bench with 'sports_club' database")
	flag.Parse()

	if (!*initFlag && !*testFlag) {
		flag.Usage()
		os.Exit(1)
	}

	if (*initFlag && *testFlag) {
		fmt.Println("You need to choose one flag!!!!!!")
		flag.Usage()
		os.Exit(1)
	}

	if *initFlag {
		initSportsDb()
	} else {
		testSportClubDb()
	}
}

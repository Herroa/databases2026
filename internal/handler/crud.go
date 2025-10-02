package handler

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"databases2026/pkg/model"

	_ "github.com/lib/pq"
)

func nullInt(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func InitDataBase(dbSettings model.DbConnectionSettings) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbSettings.Host, dbSettings.Port, dbSettings.User,
		dbSettings.Password, dbSettings.DataBaseName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}

func DbIsExist(db *sql.DB) (bool, error) {
	const query = "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname='sports_club');"
	var exists bool
	err := db.QueryRow(query).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check database existence: %v", err)
		return false, err
	}

	return exists, err
}

// --- 1. users ---
func CreateUser(db *sql.DB, email string) (int, error) {
	var id int
	err := db.QueryRow("INSERT INTO users (email) VALUES ($1) RETURNING id", email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// --- 2. coaches ---
func CreateCoach(db *sql.DB, userID int) error {
	_, err := db.Exec("INSERT INTO coaches (user_id) VALUES ($1)", userID)
	return err
}

func DeleteCoach(db *sql.DB, userID int) error {
	_, err := db.Exec("DELETE FROM coaches WHERE user_id = $1", userID)
	return err
}

// --- 3. sports ---
func CreateSport(db *sql.DB, name string) (int, error) {
	const query = "INSERT INTO sports (name) VALUES ($1) RETURNING id"
	var id int
	err := db.QueryRow(query, name).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteSport(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM sports WHERE id = $1", id)
	return err
}

// --- 4. classes ---
func CreateClass(db *sql.DB, sportID, coachID int) (int, error) {
	const query = `
		INSERT INTO classes (sport_id, coach_id) 
		VALUES ($1, $2) RETURNING id
	`

	var id int
	err := db.QueryRow(query, sportID, coachID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteClass(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM classes WHERE id = $1", id)
	return err
}

// --- 5. rooms ---
func CreateRoom(db *sql.DB, capacity int) (int, error) {
	const query = "INSERT INTO rooms (capacity) VALUES ($1) RETURNING id"
	var id int
	err := db.QueryRow(query, capacity).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteRoom(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM rooms WHERE id = $1", id)
	return err
}

// --- 6. schedules ---
func CreateSchedule(db *sql.DB, classID, roomID int, start, end time.Time) (int, error) {
	const query = `
		INSERT INTO schedules (class_id, room_id, start_time, end_time)
		VALUES ($1, $2, $3, $4) RETURNING id
	`

	var id int
	err := db.QueryRow(query, classID, roomID, start, end).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteSchedule(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM schedules WHERE id = $1", id)
	return err
}

// --- 7. bookings ---
func CreateBooking(db *sql.DB, userID, scheduleID int) (int, error) {
	const query = `
		INSERT INTO bookings (user_id, schedule_id)
		VALUES ($1, $2) RETURNING id
	`

	var id int
	err := db.QueryRow(query, userID, scheduleID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteBooking(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM bookings WHERE id = $1", id)
	return err
}

// --- 8. memberships ---
func CreateMembership(db *sql.DB, durationDays int, price float64) (int, error) {
	const query = `
		INSERT INTO memberships (duration_days, price) 
		VALUES ($1, $2) RETURNING id
	`

	var id int
	err := db.QueryRow(query, durationDays, price).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteMembership(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM memberships WHERE id = $1", id)
	return err
}

// --- 9. user_memberships ---
func CreateUserMembership(
	db *sql.DB,
	userID int,
	membershipID int,
	started time.Time,
	ended time.Time,
) (int, error) {
	const query = `
		INSERT INTO user_memberships (user_id, membership_id, started_at, ended_at)
		VALUES ($1, $2, $3, $4) RETURNING id
	`

	var id int
	err := db.QueryRow(query, userID, membershipID, started, ended).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteUserMembership(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM user_memberships WHERE id = $1", id)
	return err
}

// --- 10. payments ---
func CreatePayment(db *sql.DB, userID int, amount float64) (int, error) {
	const query = `
		INSERT INTO payments (user_id, amount)
		VALUES ($1, $2) RETURNING id
	`

	var id int
	err := db.QueryRow(query, userID, amount).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeletePayment(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM payments WHERE id = $1", id)
	return err
}

// --- 11. attendance_logs ---
func CreateAttendanceLog(db *sql.DB, userID int, start, end time.Time) (int, error) {
	const query = `
		INSERT INTO attendance_logs (user_id, start_time, end_time)
		VALUES ($1, $2, $3) RETURNING id
	`

	var id int
	err := db.QueryRow(query, userID, start, end).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteAttendanceLog(db *sql.DB, id int) error {
	const query = "DELETE FROM attendance_logs WHERE id = $1"
	_, err := db.Exec(query, id)
	return err
}

// --- 12. reviews ---
func CreateReview(db *sql.DB, userID, coachID, classID int, rating int) (int, error) {
	const query = `
		INSERT INTO reviews (user_id, coach_id, class_id, rating)
		VALUES ($1, $2, $3, $4) RETURNING id
	`

	var id int
	err := db.QueryRow(query, userID, nullInt(coachID), nullInt(classID), rating).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteReview(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM reviews WHERE id = $1", id)
	return err
}

// --- 13. promotions ---
func CreatePromotion(
	db *sql.DB,
	code string,
	discount int,
	from time.Time,
	until time.Time,
	maxUses *int,
) (int, error) {
	var id int
	if maxUses == nil {
		err := db.QueryRow(`
			INSERT INTO promotions 
			(code, discount_percent, valid_from, valid_until)
			VALUES ($1, $2, $3, $4) RETURNING id
		`, code, discount, from, until).Scan(&id)
		if err != nil {
			return 0, err
		}

		return id, err
	}

	err := db.QueryRow(`
		INSERT INTO promotions 
		(code, discount_percent, valid_from, valid_until, max_uses)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, code, discount, from, until, *maxUses).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeletePromotion(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM promotions WHERE id = $1", id)
	return err
}

// --- 14. promotion_usage ---
func UsePromotion(db *sql.DB, userID, promoID int) (int, error) {
	const query = `
		INSERT INTO promotion_usage (user_id, promotion_id) 
		VALUES ($1, $2) RETURNING id
	`

	var id int
	err := db.QueryRow(query, userID, promoID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeletePromotionUsage(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM promotion_usage WHERE id = $1", id)
	return err
}

// --- 15. notifications ---
func CreateNotification(db *sql.DB, userID int, isRead bool) (int, error) {
	const query = `
		INSERT INTO notifications (user_id, is_read) 
		VALUES ($1, $2) RETURNING id
	`

	var id int
	err := db.QueryRow(query, userID, isRead).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteNotification(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM notifications WHERE id = $1", id)
	return err
}

// --- 16. loyalty_points ---
func SetLoyaltyPoints(db *sql.DB, userID, points int) error {
	const query = `
		INSERT INTO loyalty_points (user_id, points)
		VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET points = $2
	`

	_, err := db.Exec(query,userID, points)
	return err
}

func DeleteLoyaltyPoints(db *sql.DB, userID int) error {
	_, err := db.Exec("DELETE FROM loyalty_points WHERE user_id = $1", userID)
	return err
}

// --- 17. referrals ---
func CreateReferral(db *sql.DB, referrerID, referredID int) (int, error) {
	const query = `
		INSERT INTO referrals (referrer_id, referred_id) 
		VALUES ($1, $2) RETURNING id
	`

	var id int
	err := db.QueryRow(query, referrerID, referredID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteReferral(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM referrals WHERE id = $1", id)
	return err
}

// --- 18. audit_logs ---
func LogAudit(
	db *sql.DB,
	userID *int,
	action string,
	entityType string,
	entityID *int,
) (int, error) {
	checkAndExit := func(id int, err error) (int, error) {
		if err != nil {
			return 0, err
		}
		return id, nil
	}

	var id int
	if userID == nil && entityID == nil {
		err := db.QueryRow(`
			INSERT INTO audit_logs (action) VALUES ($1) RETURNING id
		`, action).Scan(&id)
		return checkAndExit(id, err)
	}

	if userID == nil {
		err := db.QueryRow(`
			INSERT INTO audit_logs (action, entity_type, entity_id) 
			VALUES ($1, $2, $3) RETURNING id
		`, action, entityType, entityID).Scan(&id)
		return checkAndExit(id, err)
	}

	if entityID == nil {
		err := db.QueryRow(`
			INSERT INTO audit_logs (user_id, action, entity_type) 
			VALUES ($1, $2, $3) RETURNING id
		`, *userID, action, entityType).Scan(&id)
		return checkAndExit(id, err)
	}

	err := db.QueryRow(`
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id) 
		VALUES ($1, $2, $3, $4) RETURNING id
	`, *userID, action, entityType, *entityID).Scan(&id)
	return checkAndExit(id, err)
}

func DeleteAuditLog(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM audit_logs WHERE id = $1", id)
	return err
}

// --- 19. system_settings ---
func SetSystemSetting(db *sql.DB, key, value string) error {
	const query = `
		INSERT INTO system_settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $
	`

	_, err := db.Exec(query, key, value)
	return err
}

func DeleteSystemSetting(db *sql.DB, key string) error {
	_, err := db.Exec("DELETE FROM system_settings WHERE key = $1", key)
	return err
}

// --- 20. temp_bookings ---
func CreateTempBooking(
	db *sql.DB,
	userID int,
	scheduleID int,
	expires time.Time,
	token string,
) (int, error) {
	const query = `
		INSERT INTO temp_bookings (user_id, schedule_id, expires_at, token)
		VALUES ($1, $2, $3, $4) RETURNING id
	`

	var id int
	err := db.QueryRow(query, userID, scheduleID, expires, token).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func DeleteTempBooking(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM temp_bookings WHERE id = $1", id)
	return err
}

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"databases2026/internal/handler"
	"databases2026/internal/service"
	"databases2026/pkg/model"

	_ "github.com/lib/pq"
)

/*
	НАСТРОЙКИ ПОД DOCKER COMPOSE

	pg-master:
	  ports:
	    - "5433:5432"
*/

var commonDb = model.DbConnectionSettings{
	Host:         "localhost",
	Port:         5433, // ⬅️ pg-master
	User:         "postgres",
	Password:     "postgres",
	DataBaseName: "postgres",
}

var sportsDb = model.DbConnectionSettings{
	Host:         "localhost",
	Port:         5433, // ⬅️ pg-master
	User:         "postgres",
	Password:     "postgres",
	DataBaseName: "sports_club",
}

func crudTests(db *sql.DB) {
	userID, err := handler.CreateUser(db, "test78@example.com")
	if err != nil {
		log.Fatal("CreateUser:", err)
	}

	err = handler.CreateCoach(db, userID)
	if err != nil {
		log.Fatal("CreateCoach:", err)
	}

	sportID, err := handler.CreateSport(db, "Pilates")
	if err != nil {
		log.Fatal("CreateSport:", err)
	}

	classID, err := handler.CreateClass(db, sportID, userID)
	if err != nil {
		log.Fatal("CreateClass:", err)
	}

	roomID, err := handler.CreateRoom(db, 20)
	if err != nil {
		log.Fatal("CreateRoom:", err)
	}

	schedID, err := handler.CreateSchedule(
		db,
		classID,
		roomID,
		time.Now(),
		time.Now().Add(time.Hour),
	)
	if err != nil {
		log.Fatal("CreateSchedule:", err)
	}

	bookingID, err := handler.CreateBooking(db, userID, schedID)
	if err != nil {
		log.Fatal("CreateBooking:", err)
	}

	fmt.Printf("✅ Созданы сущности: user=%d, booking=%d\n", userID, bookingID)

	// Очистка
	handler.DeleteBooking(db, bookingID)
	handler.DeleteSchedule(db, schedID)
	handler.DeleteClass(db, classID)
	handler.DeleteSport(db, sportID)
	handler.DeleteCoach(db, userID)
	handler.DeleteUser(db, userID)
}

func businessCases(db *sql.DB) {
	fmt.Println("\n📊 Агрегирующие запросы")
	fmt.Printf("Общий доход: $%.2f\n", service.GetTotalRevenue(db))
	fmt.Printf("Средний рейтинг: %.2f\n", service.GetAvgClassRating(db))
	service.GetBookingsPerDay(db)
	service.GetTopSportsByAttendance(db)

	fmt.Println("\n🪟 Оконные функции")
	service.GetUserRankByLoyalty(db)
	service.GetRunningTotalRevenue(db)
	service.GetClassBookingsWithMovingAvg(db)
	service.GetCoachRatingWithRowNumber(db)

	fmt.Println("\n🔗 JOIN-запросы")
	service.GetUsersWithLoyalty(db)
	service.GetActiveMemberships(db)
	service.GetBookingsWithDetails(db)
	service.GetPaymentsWithMembership(db)
	service.GetReviewsWithCoachInfo(db)
	service.GetReferralRewards(db)
	service.GetScheduleWithRoomAndSport(db)
	service.GetFullBookingInfo(db)
}

func testSportClubDb() {
	db, err := handler.InitDataBase(sportsDb)
	if err != nil {
		log.Fatal("InitDataBase:", err)
	}
	defer db.Close()

	exists, err := handler.DbIsExist(db)
	if err != nil {
		log.Fatal("DbIsExist:", err)
	}

	if !exists {
		log.Fatal("❌ database sports_club does not exist")
	}

	fmt.Println("✅ Подключение к sports_club")

	crudTests(db)
	businessCases(db)
}

func initSportsDb() {
	db, err := handler.InitDataBase(commonDb)
	if err != nil {
		log.Fatal("InitDataBase:", err)
	}
	defer db.Close()

	fmt.Println("✅ Подключение к postgres")

	exists, err := handler.DbIsExist(db)
	if err != nil {
		log.Fatal("DbIsExist:", err)
	}

	if exists {
		log.Fatal("❌ database sports_club already exists")
	}

	_, err = db.Exec(`
		CREATE DATABASE sports_club
		ENCODING 'UTF8'
		LC_COLLATE 'en_US.UTF-8'
		LC_CTYPE 'en_US.UTF-8'
		TEMPLATE template0;
	`)
	if err != nil {
		log.Fatal("CREATE DATABASE:", err)
	}

	fmt.Println("✅ database sports_club created")

	dbSports, err := handler.InitDataBase(sportsDb)
	if err != nil {
		log.Fatal("InitDataBase:", err)
	}
	defer dbSports.Close()

	execSQL := func(path string) {
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("read %s: %v", path, err)
		}
		if _, err := dbSports.Exec(string(content)); err != nil {
			log.Fatalf("exec %s: %v", path, err)
		}
		fmt.Println("✅ executed:", path)
	}

	execSQL("../configs/sql/init_db.sql")
	execSQL("../configs/sql/generate_3m_bookings.sql")
}

func main() {
	initFlag := flag.Bool("init", false, "init sports_club database")
	testFlag := flag.Bool("test", false, "run tests and queries")
	flag.Parse()

	if *initFlag == *testFlag {
		fmt.Println("❌ choose exactly one flag: -init or -test")
		flag.Usage()
		os.Exit(1)
	}

	if *initFlag {
		initSportsDb()
	} else {
		testSportClubDb()
	}
}
package main

import (
	"database/sql"
	"databases2026/internal/handler"
	"databases2026/internal/service"
	"databases2026/pkg/model"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var commonDb = model.DbConnectionSettings{
	Host:         "localhost",
	Port:         5432,
	User:         "postgres",
	Password:     "postgres",
	DataBaseName: "postgres",
}

var sportsDb = model.DbConnectionSettings{
	Host:         "localhost",
	Port:         5432,
	User:         "postgres",
	Password:     "postgres",
	DataBaseName: "sports_club",
}

func crudTests(db *sql.DB) {
	userID, err := handler.CreateUser(db, "test78@example.com")
	if err != nil {
		fmt.Println("CreateUser: ", err)
		os.Exit(1)
	}

	err = handler.CreateCoach(db, userID)
	if err != nil {
		fmt.Println("CreateCoach: ", err)
		os.Exit(1)
	}

	sportID, err := handler.CreateSport(db, "Pilates")
	if err != nil {
		fmt.Println("CreateSport: ", err)
		os.Exit(1)
	}

	classID, err := handler.CreateClass(db, sportID, userID)
	if err != nil {
		fmt.Println("CreateClass: ", err)
		os.Exit(1)
	}

	roomID, err := handler.CreateRoom(db, 20)
	if err != nil {
		fmt.Println("CreateRoom: ", err)
		os.Exit(1)
	}

	schedID, err := handler.CreateSchedule(
		db, classID, roomID, time.Now(), time.Now().Add(time.Hour))
	if err != nil {
		fmt.Println("CreateSchedule: ", err)
		os.Exit(1)
	}

	bookingID, err := handler.CreateBooking(db, userID, schedID)
	if err != nil {
		fmt.Println("CreateBooking: ", err)
		os.Exit(1)
	}

	fmt.Printf("Созданы сущности: user=%d, booking=%d\n", userID, bookingID)

	// Очистка
	handler.DeleteBooking(db, bookingID)
	handler.DeleteSchedule(db, schedID)
	handler.DeleteClass(db, classID)
	handler.DeleteSport(db, sportID)
	handler.DeleteCoach(db, userID)
	handler.DeleteUser(db, userID)
}

func businessCases(db *sql.DB) {
	fmt.Println("\n📊 Агрегирующие:")
	fmt.Printf("Общий доход: $%.2f\n", service.GetTotalRevenue(db))
	fmt.Printf("Средний рейтинг: %.2f\n", service.GetAvgClassRating(db))
	service.GetBookingsPerDay(db)
	service.GetTopSportsByAttendance(db)

	fmt.Println("\n🪟 Оконные функции:")
	service.GetUserRankByLoyalty(db)
	service.GetRunningTotalRevenue(db)
	service.GetClassBookingsWithMovingAvg(db)
	service.GetCoachRatingWithRowNumber(db)

	fmt.Println("\n🔗 JOIN-запросы:")
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
		fmt.Println("InitDataBase:", err)
		os.Exit(1)
	}
	defer db.Close()

	isExist, err := handler.DbIsExist(db)
	if err != nil {
		fmt.Println("DbIsExist:", err)
		os.Exit(1)
	}

	if !isExist {
		log.Fatalf("'sports_club' database doesn't exists")
		os.Exit(1)
	}

	fmt.Println("✅ Подключено к 'sports_club' БД")

	crudTests(db)
	businessCases(db)
}

func initSportsDb() {
	db, err := handler.InitDataBase(commonDb)
	if err != nil {
		fmt.Println("InitDataBase:", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("✅ Подключено к 'postgres' БД")

	isExist, err := handler.DbIsExist(db)
	if err != nil {
		fmt.Println("DbIsExist:", err)
		os.Exit(1)
	}

	if isExist {
		log.Fatalf("'sports_club' database already exists")
		os.Exit(1)
	}

	// Создаём базу данных
	_, err = db.Exec(`
		CREATE DATABASE sports_club 
		ENCODING 'UTF8' 
		LC_COLLATE 'en_US.UTF-8' 
		LC_CTYPE 'en_US.UTF-8' 
		TEMPLATE template0;
	`)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
		os.Exit(1)
	}
	fmt.Println("Database sports_club created")

	// Подключаемся к базе sports_club
	dbSportsClub, err := handler.InitDataBase(sportsDb)
	if err != nil {
		fmt.Println("InitDataBase:", err)
		os.Exit(1)
	}
	defer dbSportsClub.Close()
	fmt.Println("✅ Подключено к 'sports_club' БД")

	// Функция для выполнения SQL из файла
	execSQLFile := func(filename string) {
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filename, err)
			os.Exit(1)
		}
		_, err = dbSportsClub.Exec(string(content))
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
	benchmarkFlag := flag.Bool("benchmark", false, "Run performance benchmarks")

	flag.Parse()

	flags := 0
	if *initFlag {
		flags++
	}
	if *testFlag {
		flags++
	}
	if *benchmarkFlag {
		flags++
	}

	if flags == 0 {
		fmt.Println("Error: no flag provided")
		flag.Usage()
		os.Exit(1)
	}
	if flags > 1 {
		fmt.Println("Error: only one flag allowed")
		flag.Usage()
		os.Exit(1)
	}

	if *initFlag {
		initSportsDb()
	} else if *testFlag {
		testSportClubDb()
	} else if *benchmarkFlag {
		// 🔥 НОВЫЙ БЛОК: запуск бенчмарка
		db, err := handler.InitDataBase(sportsDb)
		if err != nil {
			fmt.Println("InitDataBase:", err)
			os.Exit(1)
		}
		defer db.Close()

		isExist, err := handler.DbIsExist(db)
		if err != nil || !isExist {
			log.Fatalf("'sports_club' database doesn't exist or connection failed")
		}

		fmt.Println("✅ Connected to 'sports_club' for benchmark")
		service.RunBenchmark(db)
		fmt.Println("✅ Benchmark completed")
	}
}

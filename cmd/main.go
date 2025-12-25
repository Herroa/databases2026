package main

import (
	"context"
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
	"github.com/redis/go-redis/v9"
)

/*
	НАСТРОЙКИ ПОД DOCKER COMPOSE

	pg-master:
	  ports:
	    - "5433:5432"
*/

var commonDb = model.DbConnectionSettings{
	Host:         "localhost",
	Port:         5433,
	User:         "postgres",
	Password:     "postgres",
	DataBaseName: "postgres",
}

var sportsDb = model.DbConnectionSettings{
	Host:         "localhost",
	Port:         5433,
	User:         "postgres",
	Password:     "postgres",
	DataBaseName: "sports_club",
}

func measure(title string, fn func()) {
	start := time.Now()
	fn()
	fmt.Printf("⏱ %s took %v\n", title, time.Since(start))
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

func businessCases(db *sql.DB, rdb redis.UniversalClient, ctx context.Context) {
	fmt.Println("\n📊 Агрегирующие запросы")

	total, _ := service.GetTotalRevenue(db, rdb, ctx)
	fmt.Printf("Общий доход: $%.2f\n", total)

	avg, _ := service.GetAvgClassRating(db, rdb, ctx)
	fmt.Printf("Средний рейтинг: %.2f\n", avg)

	service.GetBookingsPerDay(db, rdb, ctx)
	service.GetTopSportsByAttendance(db, rdb, ctx)
}

func testSportClubDb() {
	db, err := handler.InitDataBase(sportsDb)
	if err != nil {
		log.Fatal("InitDataBase:", err)
	}
	defer db.Close()

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7002",
			"localhost:7003",
			"localhost:7004",
			"localhost:7005",
			"localhost:7006",
			"localhost:7007",
		},
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("cannot connect to Redis:", err)
	}

	defer rdb.Close()

	exists, err := handler.DbIsExist(db)
	if err != nil {
		log.Fatal("DbIsExist:", err)
	}

	if !exists {
		log.Fatal("❌ database sports_club does not exist")
	}

	fmt.Println("✅ Подключение к sports_club")

	crudTests(db)

	fmt.Println("\n🚀 Первый запуск (Redis MISS → Postgres)")
	measure("businessCases #1", func() {
		businessCases(db, rdb, ctx)
	})

	fmt.Println("\n⚡ Второй запуск (Redis HIT)")
	measure("businessCases #2", func() {
		businessCases(db, rdb, ctx)
	})
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

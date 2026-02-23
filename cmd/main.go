package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	_ "github.com/lib/pq"
)

const (
	mongoURI = "mongodb://localhost:27017"
	mongoDB  = "sport_club"
)

var (
	commonDb = DbConnectionSettings{
		Host:         "localhost",
		Port:         5433,
		User:         "postgres",
		Password:     "postgres",
		DataBaseName: "postgres",
	}

	sportsDb = DbConnectionSettings{
		Host:         "localhost",
		Port:         5433,
		User:         "postgres",
		Password:     "postgres",
		DataBaseName: "sports_club",
	}
)

type DbConnectionSettings struct {
	Host, User, Password, DataBaseName string
	Port                               int
}

func main() {
	rand.Seed(time.Now().UnixNano())

	pgInit := flag.Bool("pg-init", false, "init postgres")
	pgTest := flag.Bool("pg-test", false, "test postgres")
	mongoInitFlag := flag.Bool("mongo-init", false, "init mongo")
	mongoTestFlag := flag.Bool("mongo-test", false, "test mongo")
	mongoCRUDFlag := flag.Bool("mongo-crud", false, "run mongo CRUD demo")

	flag.Parse()

	switch {
	case *pgInit:
		initSportsDb()
	case *pgTest:
		testSportClubDb()
	case *mongoInitFlag:
		initMongo()
	case *mongoTestFlag:
		testMongo()
	case *mongoCRUDFlag:
		runMongoCRUD()
	default:
		fmt.Println("❌ choose a flag: -pg-init | -pg-test | -mongo-init | -mongo-test | -mongo-crud")
		os.Exit(1)
	}
}

// ====================== POSTGRES ======================

func connectPG(dbCfg DbConnectionSettings) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DataBaseName)
	return sql.Open("postgres", dsn)
}

func initSportsDb() {
	db, err := connectPG(commonDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE DATABASE sports_club;`)
	if err != nil {
		log.Fatal("Database already exists or error:", err)
	}
	fmt.Println("✅ Postgres sports_club created")
}

func testSportClubDb() {
	db, err := connectPG(sportsDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to sports_club")

	// Redis demo
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"localhost:7002", "localhost:7003", "localhost:7004", "localhost:7005", "localhost:7006", "localhost:7007"},
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("cannot connect to Redis:", err)
	}
	defer rdb.Close()
	fmt.Println("✅ Connected to Redis")
}

// ====================== MONGO ======================

func connectMongo() *mongo.Client {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func initMongo() {
	client := connectMongo()
	defer client.Disconnect(context.Background())

	db := client.Database(mongoDB)
	collections := []string{"users", "classes", "attendance_logs"}
	for _, c := range collections {
		_ = db.CreateCollection(context.Background(), c)
		fmt.Println("✅ created collection:", c)
	}
	fmt.Println("🎉 MongoDB init done")
}

func testMongo() {
	client := connectMongo()
	defer client.Disconnect(context.Background())

	db := client.Database(mongoDB)
	count, _ := db.Collection("attendance_logs").CountDocuments(context.Background(), bson.M{})
	fmt.Println("📊 attendance logs:", count)
}

// ====================== MONGO CRUD DEMO ======================

func runMongoCRUD() {
	ctx := context.Background()
	client := connectMongo()
	defer client.Disconnect(ctx)
	db := client.Database(mongoDB)

	users := db.Collection("users")
	classes := db.Collection("classes")
	attendance := db.Collection("attendance_logs")

	// create coach
	var coach bson.M
	err := users.FindOne(ctx, bson.M{"role": "coach"}).Decode(&coach)
	if err == mongo.ErrNoDocuments {
		res, _ := users.InsertOne(ctx, bson.M{"email": "coach@sportclub.com", "role": "coach", "loyaltyPoints": 0, "createdAt": time.Now()})
		coach = bson.M{"_id": res.InsertedID}
		fmt.Println("➕ Coach created")
	}
	coachID := coach["_id"].(primitive.ObjectID)

	// create client
	var clientUser bson.M
	err = users.FindOne(ctx, bson.M{"role": "client"}).Decode(&clientUser)
	if err == mongo.ErrNoDocuments {
		res, _ := users.InsertOne(ctx, bson.M{"email": "client@sportclub.com", "role": "client", "loyaltyPoints": 100, "createdAt": time.Now()})
		clientUser = bson.M{"_id": res.InsertedID}
		fmt.Println("➕ Client created")
	}
	clientID := clientUser["_id"].(primitive.ObjectID)

	// insert user
	resUser, _ := users.InsertOne(ctx, bson.M{"email": "new_user@mail.com", "role": "client", "loyaltyPoints": 100, "createdAt": time.Now()})
	userID := resUser.InsertedID.(primitive.ObjectID)

	// insert classes
	classes.InsertMany(ctx, []interface{}{
		bson.M{"title": "Morning Yoga", "sport": "Yoga", "coachId": coachID, "capacity": 15, "schedule": bson.M{"date": time.Now(), "startTime": "08:00", "endTime": "09:00"}, "bookings": []interface{}{}},
		bson.M{"title": "Evening Boxing", "sport": "Boxing", "coachId": coachID, "capacity": 20, "schedule": bson.M{"date": time.Now(), "startTime": "19:00", "endTime": "20:00"}, "bookings": []interface{}{}},
	})

	// updateOne $set
	users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"email": "updated_user@mail.com"}})

	// updateOne $inc
	users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$inc": bson.M{"loyaltyPoints": 50}})

	// push booking
	classes.UpdateOne(ctx, bson.M{"title": "Morning Yoga"}, bson.M{"$push": bson.M{"bookings": bson.M{"userId": clientID, "bookedAt": time.Now()}}})

	// updateMany
	users.UpdateMany(ctx, bson.M{"role": "client"}, bson.M{"$set": bson.M{"loyaltyPoints": 0}})

	// deleteMany
	attendance.DeleteMany(ctx, bson.M{"attendedAt": bson.M{"$lt": time.Now().Add(-30 * 24 * time.Hour)}})

	// upsert
	users.UpdateOne(ctx, bson.M{"email": "upsert_user@mail.com"}, bson.M{"$set": bson.M{"role": "client", "loyaltyPoints": 50, "createdAt": time.Now()}}, options.Update().SetUpsert(true))

	fmt.Println("✅ MongoDB CRUD demo completed")
}

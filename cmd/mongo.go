package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoURI = "mongodb://localhost:27017"
	dbName   = "sport_club"
)

var ctx = context.Background()

// ---------- CONNECT ----------

func connectMongo() *mongo.Client {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ Connected to MongoDB")
	return client
}

// ---------- INIT DATABASE ----------

func initMongo() {
	client := connectMongo()
	defer client.Disconnect(ctx)

	db := client.Database(dbName)

	// collections
	collections := []string{"users", "classes", "attendance_logs"}
	for _, c := range collections {
		err := db.CreateCollection(ctx, c)
		if err == nil {
			fmt.Println("✅ created collection:", c)
		}
	}

	seedUsers(db)
	seedClasses(db)
	seedAttendance(db)

	fmt.Println("🎉 MongoDB seed completed")
}

// ---------- SEED USERS ----------

func seedUsers(db *mongo.Database) {
	db.Collection("users").DeleteMany(ctx, bson.M{})

	var users []interface{}

	// coaches
	for i := 1; i <= 20; i++ {
		users = append(users, bson.M{
			"email":          fmt.Sprintf("coach%d@sportclub.com", i),
			"role":           "coach",
			"loyaltyPoints":  0,
			"createdAt":      time.Now(),
		})
	}

	// clients
	for i := 1; i <= 300; i++ {
		users = append(users, bson.M{
			"email":         fmt.Sprintf("user%d@mail.com", i),
			"role":          "client",
			"loyaltyPoints": rand.Intn(500),
			"createdAt":     time.Now().AddDate(0, 0, -rand.Intn(180)),
		})
	}

	db.Collection("users").InsertMany(ctx, users)
	fmt.Println("✅ users seeded:", len(users))
}

// ---------- SEED CLASSES ----------

func seedClasses(db *mongo.Database) {
	db.Collection("classes").DeleteMany(ctx, bson.M{})

	coaches, _ := db.Collection("users").Find(ctx, bson.M{"role": "coach"})
	clients, _ := db.Collection("users").Find(ctx, bson.M{"role": "client"})

	var coachIDs []primitive.ObjectID
	var clientIDs []primitive.ObjectID

	for coaches.Next(ctx) {
		var u bson.M
		coaches.Decode(&u)
		coachIDs = append(coachIDs, u["_id"].(primitive.ObjectID))
	}

	for clients.Next(ctx) {
		var u bson.M
		clients.Decode(&u)
		clientIDs = append(clientIDs, u["_id"].(primitive.ObjectID))
	}

	sports := []string{"Yoga", "Boxing", "Crossfit", "Pilates", "Swimming"}

	var classes []interface{}

	for i := 0; i < 50; i++ {
		bookingsCount := rand.Intn(20)
		var bookings []bson.M

		for j := 0; j < bookingsCount; j++ {
			bookings = append(bookings, bson.M{
				"userId":   clientIDs[rand.Intn(len(clientIDs))],
				"bookedAt": time.Now(),
			})
		}

		classes = append(classes, bson.M{
			"title":    fmt.Sprintf("Class #%d", i+1),
			"sport":    sports[i%len(sports)],
			"coachId":  coachIDs[i%len(coachIDs)],
			"capacity": 20,
			"schedule": bson.M{
				"date":      time.Now().AddDate(0, 0, i),
				"startTime": "10:00",
				"endTime":   "11:00",
			},
			"bookings": bookings,
		})
	}

	db.Collection("classes").InsertMany(ctx, classes)
	fmt.Println("✅ classes seeded:", len(classes))
}

// ---------- SEED ATTENDANCE ----------

func seedAttendance(db *mongo.Database) {
	db.Collection("attendance_logs").DeleteMany(ctx, bson.M{})

	clients, _ := db.Collection("users").Find(ctx, bson.M{"role": "client"})
	classes, _ := db.Collection("classes").Find(ctx, bson.M{})

	var clientIDs []primitive.ObjectID
	var classIDs []primitive.ObjectID

	for clients.Next(ctx) {
		var u bson.M
		clients.Decode(&u)
		clientIDs = append(clientIDs, u["_id"].(primitive.ObjectID))
	}

	for classes.Next(ctx) {
		var c bson.M
		classes.Decode(&c)
		classIDs = append(classIDs, c["_id"].(primitive.ObjectID))
	}

	var logs []interface{}

	for i := 0; i < 600; i++ {
		logs = append(logs, bson.M{
			"userId":     clientIDs[rand.Intn(len(clientIDs))],
			"classId":    classIDs[rand.Intn(len(classIDs))],
			"attendedAt": time.Now().AddDate(0, 0, -rand.Intn(60)),
		})
	}

	db.Collection("attendance_logs").InsertMany(ctx, logs)
	fmt.Println("✅ attendance_logs seeded:", len(logs))
}

// ---------- TEST / DEMO ----------

func testMongo() {
	client := connectMongo()
	defer client.Disconnect(ctx)

	db := client.Database(dbName)

	count, _ := db.Collection("attendance_logs").CountDocuments(ctx, bson.M{})
	fmt.Println("📊 attendance logs:", count)

	cur, _ := db.Collection("users").Find(ctx, bson.M{"role": "coach"})
	for cur.Next(ctx) {
		var u bson.M
		cur.Decode(&u)
		fmt.Println(" -", u["email"])
	}
}

// ---------- MAIN ----------

func main() {
	rand.Seed(time.Now().UnixNano())

	initFlag := flag.Bool("init", false, "init mongo database")
	testFlag := flag.Bool("test", false, "run demo queries")
	flag.Parse()

	if *initFlag == *testFlag {
		log.Fatal("❌ flag: -init or -test")
	}

	if *initFlag {
		initMongo()
	} else {
		testMongo()
	}
}
package main

import (
	"context"
	"fmt"
	"log"
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

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(dbName)

	users := db.Collection("users")
	classes := db.Collection("classes")
	attendance := db.Collection("attendance_logs")

	// ============================================================
	// 🔎 Получить или создать COACH
	// ============================================================
	var coach bson.M
	err = users.FindOne(ctx, bson.M{"role": "coach"}).Decode(&coach)
	if err == mongo.ErrNoDocuments {
		res, _ := users.InsertOne(ctx, bson.M{
			"email":         "coach@sportclub.com",
			"role":          "coach",
			"loyaltyPoints": 0,
			"createdAt":     time.Now(),
		})
		coach = bson.M{"_id": res.InsertedID}
		fmt.Println("➕ Coach created")
	}
	coachID := coach["_id"].(primitive.ObjectID)

	// ============================================================
	// 🔎 Получить или создать CLIENT
	// ============================================================
	var clientUser bson.M
	err = users.FindOne(ctx, bson.M{"role": "client"}).Decode(&clientUser)
	if err == mongo.ErrNoDocuments {
		res, _ := users.InsertOne(ctx, bson.M{
			"email":         "client@sportclub.com",
			"role":          "client",
			"loyaltyPoints": 100,
			"createdAt":     time.Now(),
		})
		clientUser = bson.M{"_id": res.InsertedID}
		fmt.Println("➕ Client created")
	}
	clientID := clientUser["_id"].(primitive.ObjectID)

	// ============================================================
	// insertOne — пользователь
	// ============================================================
	fmt.Println("▶ insertOne user")

	resUser, _ := users.InsertOne(ctx, bson.M{
		"email":         "new_user@mail.com",
		"role":          "client",
		"loyaltyPoints": 100,
		"createdAt":     time.Now(),
	})
	userID := resUser.InsertedID.(primitive.ObjectID)

	// ============================================================
	// insertMany — классы
	// ============================================================
	fmt.Println("▶ insertMany classes")

	classes.InsertMany(ctx, []interface{}{
		bson.M{
			"title":    "Morning Yoga",
			"sport":    "Yoga",
			"coachId":  coachID,
			"capacity": 15,
			"schedule": bson.M{
				"date":      time.Now(),
				"startTime": "08:00",
				"endTime":   "09:00",
			},
			"bookings": []interface{}{},
		},
		bson.M{
			"title":    "Evening Boxing",
			"sport":    "Boxing",
			"coachId":  coachID,
			"capacity": 20,
			"schedule": bson.M{
				"date":      time.Now(),
				"startTime": "19:00",
				"endTime":   "20:00",
			},
			"bookings": []interface{}{},
		},
	})

	// ============================================================
	// updateOne + $set
	// ============================================================
	fmt.Println("▶ updateOne $set")

	users.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"email": "updated_user@mail.com"}},
	)

	// ============================================================
	// updateOne + $inc
	// ============================================================
	fmt.Println("▶ updateOne $inc")

	users.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$inc": bson.M{"loyaltyPoints": 50}},
	)

	// ============================================================
	// $push booking
	// ============================================================
	fmt.Println("▶ $push booking")

	classes.UpdateOne(
		ctx,
		bson.M{"title": "Morning Yoga"},
		bson.M{
			"$push": bson.M{
				"bookings": bson.M{
					"userId":   clientID,
					"bookedAt": time.Now(),
				},
			},
		},
	)

	// ============================================================
	// updateMany
	// ============================================================
	fmt.Println("▶ updateMany")

	users.UpdateMany(
		ctx,
		bson.M{"role": "client"},
		bson.M{"$set": bson.M{"loyaltyPoints": 0}},
	)

	// ============================================================
	// deleteMany
	// ============================================================
	fmt.Println("▶ deleteMany")

	attendance.DeleteMany(
		ctx,
		bson.M{
			"attendedAt": bson.M{
				"$lt": time.Now().Add(-30 * 24 * time.Hour),
			},
		},
	)

	// ============================================================
	// upsert
	// ============================================================
	fmt.Println("▶ upsert")

	users.UpdateOne(
		ctx,
		bson.M{"email": "upsert_user@mail.com"},
		bson.M{
			"$set": bson.M{
				"role":          "client",
				"loyaltyPoints": 50,
				"createdAt":     time.Now(),
			},
		},
		options.Update().SetUpsert(true),
	)

	fmt.Println("\n✅ Все MongoDB CRUD-операции выполнены корректно")
}
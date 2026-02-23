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

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
	neo4jTestFlag := flag.Bool("neo4j-test", false, "test neo4j")

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
	case *neo4jTestFlag:
		testNeo4j()
	default:
		fmt.Println("❌ choose a flag: -pg-init | -pg-test | -mongo-init | -mongo-test | -mongo-crud | -neo4j-test")
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

// ====================== NEO4J =============================
func testNeo4j() {
	driver, err := neo4j.NewDriverWithContext(
		"neo4j://localhost:7687",
		neo4j.BasicAuth("neo4j", "password", ""),
	)
	if err != nil {
		log.Fatalf("❌ Cannot create Neo4j driver: %v", err)
	}
	defer driver.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("❌ Cannot connect to Neo4j: %v", err)
	}
	fmt.Println("✅ Connected to Neo4j")

	session := driver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(context.Background())

	session.ExecuteWrite(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
		return nil, err
	})

	// ====================== Seeding ======================
	seedNeo4j(ctx, session)

	// ====================== Базовые запросы ======================

	// Какие тренировки посещает клиент?
	queryClientClasses := func(clientName string) {
		fmt.Printf("\n📌 Classes for client %s:\n", clientName)
		titles, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(context.Background(),
				`MATCH (c:Client {name:$name})-[:BOOKED]->(cls:Class) RETURN cls.title`,
				map[string]any{"name": clientName})
			if err != nil {
				return nil, err
			}
			var out []string
			for result.Next(context.Background()) {
				out = append(out, result.Record().Values[0].(string))
			}
			return out, nil
		})
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(titles)
	}

	// Кто является тренером группы?
	queryClassCoach := func(classTitle string) {
		fmt.Printf("\n📌 Coach for class %s:\n", classTitle)
		coaches, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(context.Background(),
				`MATCH (cls:Class {title:$title})-[:TAUGHT_BY]->(t:Coach) RETURN t.name`,
				map[string]any{"title": classTitle})
			if err != nil {
				return nil, err
			}
			var out []string
			for result.Next(context.Background()) {
				out = append(out, result.Record().Values[0].(string))
			}
			return out, nil
		})
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(coaches)
	}

	// Какие тренировки проходят в зале
	queryClassesInRoom := func(roomName string) {
		fmt.Printf("\n📌 Classes in room %s:\n", roomName)
		classes, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(context.Background(),
				`MATCH (cls:Class)-[:HELD_IN]->(r:Room {name:$room}) RETURN cls.title`,
				map[string]any{"room": roomName})
			if err != nil {
				return nil, err
			}
			var out []string
			for result.Next(context.Background()) {
				out = append(out, result.Record().Values[0].(string))
			}
			return out, nil
		})
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(classes)
	}

	// Какие клиенты записаны на одно занятие
	queryClientsInClass := func(classTitle string) {
		fmt.Printf("\n📌 Clients in class %s:\n", classTitle)
		clients, err := session.ExecuteRead(context.Background(), func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(context.Background(),
				`MATCH (c:Client)-[:BOOKED]->(cls:Class {title:$title}) RETURN c.name`,
				map[string]any{"title": classTitle})
			if err != nil {
				return nil, err
			}
			var out []string
			for result.Next(context.Background()) {
				out = append(out, result.Record().Values[0].(string))
			}
			return out, nil
		})
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(clients)
	}

	// 🔹 Клиенты, посещающие одни и те же занятия
	queryClientsSharedClasses := func() {
		fmt.Println("\n📌 Clients attending the same classes:")

		_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, `
			MATCH (c1:Client)-[:BOOKED]->(cls:Class)<-[:BOOKED]-(c2:Client)
			WHERE c1.name < c2.name
			RETURN c1.name, c2.name, collect(cls.title) AS sharedClasses
			LIMIT 20
		`, nil)
			if err != nil {
				return nil, err
			}

			for result.Next(ctx) {
				record := result.Record()
				fmt.Printf(" - %s & %s share classes: %v\n",
					record.Values[0], record.Values[1], record.Values[2])
			}
			return nil, nil
		})
		if err != nil {
			log.Println(err)
		}
	}

	// 🔹 Тренеры, ведущие одинаковые виды занятий
	queryCoachesSameSport := func() {
		fmt.Println("\n📌 Coaches teaching the same sport:")

		_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, `
			MATCH (c1:Coach)<-[:TAUGHT_BY]-(cls:Class)-[:TAUGHT_BY]->(c2:Coach)
			WHERE c1.name < c2.name
			RETURN c1.name, c2.name, collect(cls.sport) AS commonSports
			LIMIT 20
		`, nil)
			if err != nil {
				return nil, err
			}

			for result.Next(ctx) {
				record := result.Record()
				fmt.Printf(" - %s & %s teach: %v\n",
					record.Values[0], record.Values[1], record.Values[2])
			}
			return nil, nil
		})
		if err != nil {
			log.Println(err)
		}
	}

	// 🔹 Клиенты, посещающие занятия одного тренера
	queryClientsSameCoach := func() {
		fmt.Println("\n📌 Clients attending classes of the same coach:")

		_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, `
			MATCH (c1:Client)-[:BOOKED]->(:Class)-[:TAUGHT_BY]->(coach:Coach)<-[:TAUGHT_BY]-(:Class)<-[:BOOKED]-(c2:Client)
			WHERE c1.name < c2.name
			RETURN c1.name, c2.name, collect(DISTINCT coach.name) AS sharedCoaches
			LIMIT 20
		`, nil)
			if err != nil {
				return nil, err
			}

			for result.Next(ctx) {
				record := result.Record()
				fmt.Printf(" - %s & %s share coaches: %v\n",
					record.Values[0], record.Values[1], record.Values[2])
			}
			return nil, nil
		})
		if err != nil {
			log.Println(err)
		}
	}

	// 🔹 Через какие тренировки связаны два клиента
	querySharedClassesBetweenClients := func(client1, client2 string) {
		fmt.Printf("\n📌 Classes connecting %s and %s:\n", client1, client2)
		_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, `
			MATCH (c1:Client {name:$c1})-[:BOOKED]->(cls:Class)<-[:BOOKED]-(c2:Client {name:$c2})
			RETURN cls.title AS class
		`, map[string]any{"c1": client1, "c2": client2})
			if err != nil {
				return nil, err
			}

			var classes []string
			for result.Next(ctx) {
				classes = append(classes, result.Record().Values[0].(string))
			}
			fmt.Println(classes)
			return nil, nil
		})
		if err != nil {
			log.Println(err)
		}
	}

	// 🔹 Через каких тренеров можно связать двух клиентов
	querySharedCoachesBetweenClients := func(client1, client2 string) {
		fmt.Printf("\n📌 Coaches connecting %s and %s:\n", client1, client2)
		_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, `
			MATCH (c1:Client {name:$c1})-[:BOOKED]->(:Class)-[:TAUGHT_BY]->(coach:Coach)<-[:TAUGHT_BY]-(:Class)<-[:BOOKED]-(c2:Client {name:$c2})
			RETURN DISTINCT coach.name AS coach
		`, map[string]any{"c1": client1, "c2": client2})
			if err != nil {
				return nil, err
			}

			var coaches []string
			for result.Next(ctx) {
				coaches = append(coaches, result.Record().Values[0].(string))
			}
			fmt.Println(coaches)
			return nil, nil
		})
		if err != nil {
			log.Println(err)
		}
	}

	// 🔹 Кратчайший путь между двумя залами через занятия и тренеров
	queryShortestPathBetweenRooms := func(room1, room2 string) {
		fmt.Printf("\n📌 Shortest path between rooms %s and %s:\n", room1, room2)
		_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, `
			MATCH (r1:Room {name:$r1}), (r2:Room {name:$r2})
			MATCH p = (r1)<-[:HELD_IN]-(:Class)-[:TAUGHT_BY]->(:Coach)<-[:TAUGHT_BY]-(:Class)-[:HELD_IN]->(r2)
			RETURN [n IN nodes(p) |
				CASE
					WHEN n:Room THEN n.name
					WHEN n:Class THEN n.title
					WHEN n:Coach THEN n.name
					ELSE n.name
				END
			] AS path
			LIMIT 1
		`, map[string]any{"r1": room1, "r2": room2})
			if err != nil {
				return nil, err
			}

			for result.Next(ctx) {
				fmt.Println(result.Record().Values[0])
			}
			return nil, nil
		})
		if err != nil {
			log.Println(err)
		}
	}

	// ====================== Демонстрация ======================
	queryClientClasses("Client1")
	queryClientClasses("Client2")
	queryClassCoach("Class1")
	queryClassCoach("Class10")
	queryClassesInRoom("Room1")
	queryClassesInRoom("Room3")
	queryClientsInClass("Class1")
	queryClientsInClass("Class10")

	queryClientsSharedClasses()
	queryCoachesSameSport()
	queryClientsSameCoach()

	querySharedClassesBetweenClients("Client1", "Client5")
	querySharedCoachesBetweenClients("Client1", "Client5")
	queryShortestPathBetweenRooms("Room1", "Room3")

	// 	#### 5. Спортивный центр

	// **Базовые связи**
	// - Какие тренировки посещает клиент?
	// - Кто является тренером группы?
	// - Какие тренировки проходят в зале?
	// - Какие клиенты записаны на одно занятие?

	// **Роли**
	// - Чем отличается тренер от клиента?
	// - Может ли тренер быть клиентов?
	// - Кто является администратором зала?

	// **Аналитика**
	// - Какие клиенты посещают одни и те же тренировки?
	// - Какие тренеры ведут одинаковые типы занятий?
	// - Какие клиенты посещают занятия одного тренера?

	// **Сложные графовые связи**
	// - Через какие тренировки связаны два клиента?
	// - Через каких тренеров можно связать двух клиентов?
	// - Какой кратчайший путь между двумя залами через занятия и тренеров?

}
func seedNeo4j(ctx context.Context, session neo4j.SessionWithContext) {
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

		// ====================== ФИКСИРОВАННЫЕ СУЩНОСТИ ДЛЯ ДЕМО ======================
		_, err := tx.Run(ctx, `
			MERGE (alice:Client {name:'Alice'})
			MERGE (bob:Client {name:'Bob'})

			MERGE (john:Coach {name:'John'})
			MERGE (emma:Coach {name:'Emma'})

			MERGE (yogaRoom:Room {name:'YogaRoom'})
			MERGE (boxingRoom:Room {name:'BoxingRoom'})
			MERGE (room1:Room {name:'Room1'})
			MERGE (room3:Room {name:'Room3'})

			MERGE (morningYoga:Class {title:'Morning Yoga', sport:'Yoga'})
			MERGE (eveningBoxing:Class {title:'Evening Boxing', sport:'Boxing'})
			MERGE (class1:Class {title:'Class1', sport:'Crossfit'})
			MERGE (class2:Class {title:'Class2', sport:'Crossfit'})

			MERGE (morningYoga)-[:TAUGHT_BY]->(john)
			MERGE (eveningBoxing)-[:TAUGHT_BY]->(emma)
			MERGE (morningYoga)-[:HELD_IN]->(yogaRoom)
			MERGE (eveningBoxing)-[:HELD_IN]->(boxingRoom)

			MERGE (class1)-[:HELD_IN]->(room1)
			MERGE (class2)-[:HELD_IN]->(room3)
			MERGE (class1)-[:TAUGHT_BY]->(john)
			MERGE (class2)-[:TAUGHT_BY]->(john)

			MERGE (alice)-[:BOOKED]->(morningYoga)
			MERGE (bob)-[:BOOKED]->(eveningBoxing)
			MERGE (alice)-[:BOOKED]->(class1)
			MERGE (bob)-[:BOOKED]->(class2)
		`, nil)
		if err != nil {
			return nil, err
		}

		// ====================== СЛУЧАЙНЫЕ СУЩНОСТИ ======================
		// Клиенты 1-20
		for i := 1; i <= 20; i++ {
			q := fmt.Sprintf(`MERGE (:Client {name:'Client%d'})`, i)
			if _, err := tx.Run(ctx, q, nil); err != nil {
				return nil, err
			}
		}

		// Тренеры 1-10
		for i := 1; i <= 10; i++ {
			q := fmt.Sprintf(`MERGE (:Coach {name:'Coach%d'})`, i)
			if _, err := tx.Run(ctx, q, nil); err != nil {
				return nil, err
			}
		}

		// Залы 1-5
		for i := 1; i <= 5; i++ {
			q := fmt.Sprintf(`MERGE (:Room {name:'Room%d'})`, i)
			if _, err := tx.Run(ctx, q, nil); err != nil {
				return nil, err
			}
		}

		// Классы 1-50
		sports := []string{"Yoga", "Boxing", "Crossfit", "Pilates", "Swimming"}
		for i := 1; i <= 50; i++ {
			sport := sports[i%len(sports)]
			q := fmt.Sprintf(`MERGE (:Class {title:'Class%d', sport:'%s'})`, i, sport)
			if _, err := tx.Run(ctx, q, nil); err != nil {
				return nil, err
			}
		}

		// Связи TAUGHT_BY
		for i := 1; i <= 50; i++ {
			coachID := (i%10 + 1)
			q := fmt.Sprintf(`
				MATCH (cls:Class {title:'Class%d'}), (c:Coach {name:'Coach%d'})
				MERGE (cls)-[:TAUGHT_BY]->(c)`, i, coachID)
			if _, err := tx.Run(ctx, q, nil); err != nil {
				return nil, err
			}
		}

		// Связи HELD_IN
		for i := 1; i <= 50; i++ {
			roomID := (i%5 + 1)
			q := fmt.Sprintf(`
				MATCH (cls:Class {title:'Class%d'}), (r:Room {name:'Room%d'})
				MERGE (cls)-[:HELD_IN]->(r)`, i, roomID)
			if _, err := tx.Run(ctx, q, nil); err != nil {
				return nil, err
			}
		}

		// Связи BOOKED для клиентов 1-20 (5-10 случайных занятий)
		for i := 1; i <= 20; i++ {
			for j := 0; j < 5+rand.Intn(6); j++ {
				classID := 1 + rand.Intn(50)
				q := fmt.Sprintf(`
					MATCH (c:Client {name:'Client%d'}), (cls:Class {title:'Class%d'})
					MERGE (c)-[:BOOKED]->(cls)`, i, classID)
				if _, err := tx.Run(ctx, q, nil); err != nil {
					return nil, err
				}
			}
		}

		return nil, nil
	})
	if err != nil {
		log.Fatalf("❌ Cannot seed Neo4j: %v", err)
	}

	fmt.Println("✅ Neo4j seeded with multiple entities and 100+ relationships (fixed + random)")
}

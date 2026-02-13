import { Pool } from "pg";
import Redis from "ioredis";

// Postgres
const pg = new Pool({
  host: "127.0.0.1",
  port: 5433,
  user: "postgres",
  password: "postgres",
  database: "sports_club"
});

// Redis Cluster
const redis = new Redis.Cluster([
  { host: "127.0.0.1", port: 7002 },
  { host: "127.0.0.1", port: 7003 },
  { host: "127.0.0.1", port: 7004 },
  { host: "127.0.0.1", port: 7005 },
  { host: "127.0.0.1", port: 7006 },
  { host: "127.0.0.1", port: 7007 },
]);

async function seedRedis() {
  try {
    // 1️⃣ Users
    const usersRes = await pg.query("SELECT id, email FROM users");
    for (const user of usersRes.rows) {
      await redis.hset(`user:${user.id}`, { id: user.id, email: user.email });
    }
    console.log("Users seeded");

    // 2️⃣ Schedules
    const schedulesRes = await pg.query(`
      SELECT s.id AS schedule_id, s.start_time, s.end_time,
             c.id AS class_id, c.sport_id, c.coach_id,
             sp.name AS sport_name,
             u.email AS coach_email,
             r.id AS room_id, r.capacity AS room_capacity
      FROM schedules s
      JOIN classes c ON c.id = s.class_id
      JOIN sports sp ON sp.id = c.sport_id
      JOIN coaches co ON co.user_id = c.coach_id
      JOIN users u ON u.id = co.user_id
      JOIN rooms r ON r.id = s.room_id
    `);

    for (const row of schedulesRes.rows) {
      const key = `schedule:${row.schedule_id}`;
      await redis.hset(key, {
        id: row.schedule_id,
        class_id: row.class_id,
        sport_id: row.sport_id,
        sport_name: row.sport_name,
        coach_id: row.coach_id,
        coach_email: row.coach_email,
        room_id: row.room_id,
        room_capacity: row.room_capacity,
        start_time: row.start_time.toISOString(),
        end_time: row.end_time.toISOString()
      });
      await redis.expire(key, 300);
    }

    console.log("Schedules seeded");

    console.log("✅ Redis Cluster seeding finished");
    process.exit(0);

  } catch (err) {
    console.error("Error seeding Redis Cluster:", err);
    process.exit(1);
  }
}

seedRedis();

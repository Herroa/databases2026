const express = require("express");
const { ObjectId } = require("mongodb");
const connectDB = require("../db");

const router = express.Router();

router.get("/", async (req, res) => {
  const db = await connectDB();
  const logs = await db.collection("attendance_logs").find().limit(100).toArray();
  res.json(logs);
});

router.post("/", async (req, res) => {
  const db = await connectDB();

  const doc = {
    userId: new ObjectId(req.body.userId),
    classId: new ObjectId(req.body.classId),
    attendedAt: new Date(req.body.attendedAt)
  };

  const result = await db.collection("attendance_logs").insertOne(doc);

  res.status(201).json({
    message: "Attendance added",
    attendance: {
      _id: result.insertedId,
      ...doc
    }
  });
});

router.post("/tx", async (req, res) => {
  const client = await connectDB(true);
  const session = client.startSession();

  try {
    session.startTransaction();
    const db = client.db("sport_club");

    const userId = new ObjectId(req.body.userId);
    const classId = new ObjectId(req.body.classId);

    const cls = await db.collection("classes").findOne(
      { _id: classId },
      { session }
    );

    if (!cls) {
      throw new Error("Class not found");
    }

    if (typeof cls.capacity !== "number") {
      throw new Error("Class capacity is not defined");
    }

    const count = await db.collection("attendance_logs").countDocuments(
      { classId },
      { session }
    );

    if (count >= cls.capacity) {
      throw new Error("Class is full");
    }

    await db.collection("attendance_logs").insertOne(
      {
        userId,
        classId,
        attendedAt: new Date()
      },
      { session }
    );

    await db.collection("users").updateOne(
      { _id: userId },
      { $inc: { visitsCount: 1 } },
      { session }
    );

    await db.collection("tx_logs").insertOne(
      {
        type: "attendance_tx",
        userId,
        classId,
        createdAt: new Date(),
        status: "success"
      },
      { session }
    );

    await session.commitTransaction();

    res.json({
      message: "Attendance saved (transaction)",
      classCapacity: cls.capacity,
      currentCount: count + 1
    });

  } catch (err) {
    await session.abortTransaction();
    res.status(400).json({ error: err.message });
  } finally {
    session.endSession();
  }
});

// GET /api/attendance/tx
// просмотр транзакций в браузере
router.get("/tx", async (req, res) => {
  const db = await connectDB();

  const tx = await db
    .collection("tx_logs")
    .find({ type: "attendance_tx" })
    .sort({ createdAt: -1 })
    .limit(20)
    .toArray();

  res.json({
    message: "Attendance transactions",
    data: tx
  });
});

// POST /api/attendance/bulk
// BulkWrite demo
router.post("/bulk", async (req, res) => {
  const db = await connectDB();

  const ops = [
    // 1️⃣ updateOne — увеличить visitsCount
    {
      updateOne: {
        filter: { role: "client" },
        update: { $inc: { visitsCount: 1 } }
      }
    },

    // 2️⃣ updateMany — отметить активных пользователей
    {
      updateMany: {
        filter: { visitsCount: { $gte: 5 } },
        update: { $set: { isActive: true } }
      }
    },

    // 3️⃣ deleteMany — удалить тестовых пользователей
    {
      deleteMany: {
        filter: { email: /test/i }
      }
    },

    // 4️⃣ insertOne — лог операции
    {
      insertOne: {
        document: {
          type: "bulk_operation",
          createdAt: new Date(),
          description: "Mass update users"
        }
      }
    }
  ];

  const result = await db.collection("users").bulkWrite(ops);

  res.json({
    message: "Bulk operation completed",
    result
  });
});

// GET /api/attendance/bulk
// посмотреть результат bulk-операций в браузере
router.get("/bulk", async (req, res) => {
  const db = await connectDB();

  const result = await db
    .collection("users")
    .find(
      { type: "bulk_operation" },
      { sort: { createdAt: -1 } }
    )
    .limit(10)
    .toArray();

  res.json({
    message: "Last bulk operations",
    data: result
  });
});

module.exports = router;
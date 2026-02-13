const express = require("express");
const connectDB = require("../db");

const router = express.Router();

router.get("/", async (req, res) => {
  const db = await connectDB();
  const classes = await db.collection("classes").find().toArray();
  res.json(classes);
});

router.get("/fill", async (req, res) => {
  const db = await connectDB();

  const result = await db.collection("attendance_logs").aggregate([
    {
      $group: {
        _id: "$classId",
        attendance: { $sum: 1 }
      }
    },
    {
      $lookup: {
        from: "classes",
        localField: "_id",
        foreignField: "_id",
        as: "class"
      }
    },
    { $unwind: "$class" },
    {
      $project: {
        classTitle: "$class.title",
        attendance: 1,
        capacity: "$class.capacity",
        fillPercent: {
          $multiply: [
            { $divide: ["$attendance", "$class.capacity"] },
            100
          ]
        }
      }
    },
    { $sort: { fillPercent: -1 } }
  ]).toArray();

  res.json(result);
});

// Ежедневная статистика
// GET /api/classes/fill/daily
router.get("/fill/daily", async (req, res) => {
  const db = await connectDB();

  const filter = {};
  if (req.query.classId) {
    filter.classId = new ObjectId(req.query.classId);
  }
  if (req.query.date) {
    filter.date = req.query.date;
  }

  const data = await db
    .collection("class_daily_fill")
    .find(filter)
    .sort({ date: 1 })
    .toArray();

  res.json(data);
});

// Отчет за неделю
// GET /api/classes/fill/weekly
router.get("/fill/weekly", async (req, res) => {
  const db = await connectDB();

  const result = await db.collection("attendance_logs").aggregate([
    {
      $addFields: {
        dayOfWeek: { $dayOfWeek: "$attendedAt" }
      }
    },
    {
      $group: {
        _id: {
          classId: "$classId",
          dayOfWeek: "$dayOfWeek"
        },
        attendance: { $sum: 1 }
      }
    },
    {
      $lookup: {
        from: "classes",
        localField: "_id.classId",
        foreignField: "_id",
        as: "class"
      }
    },
    { $unwind: "$class" },
    {
      $addFields: {
        fillPercent: {
          $multiply: [
            { $divide: ["$attendance", "$class.capacity"] },
            100
          ]
        }
      }
    },
    {
      $facet: {
        byClassAndDay: [
          {
            $project: {
              _id: 0,
              class: "$class.title",
              dayOfWeek: "$_id.dayOfWeek",
              attendance: 1,
              capacity: "$class.capacity",
              fillPercent: { $round: ["$fillPercent", 2] }
            }
          },
          { $sort: { class: 1, dayOfWeek: 1 } }
        ],
        topLoaded: [
          { $sort: { fillPercent: -1 } },
          { $limit: 5 },
          {
            $project: {
              _id: 0,
              class: "$class.title",
              dayOfWeek: "$_id.dayOfWeek",
              fillPercent: { $round: ["$fillPercent", 2] }
            }
          }
        ]
      }
    }
  ]).toArray();

  res.json(result[0]);
});

module.exports = router;
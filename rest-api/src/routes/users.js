const express = require("express");
const { ObjectId } = require("mongodb");
const connectDB = require("../db");

const router = express.Router();

//
// GET /api/users
// открыть в браузере
//
router.get("/", async (req, res) => {
  const db = await connectDB();
  const users = await db.collection("users").find().toArray();
  res.json(users);
});

//
// GET /api/users/:id
//
router.get("/:id", async (req, res) => {
  const db = await connectDB();
  const user = await db.collection("users").findOne({
    _id: new ObjectId(req.params.id)
  });

  if (!user) {
    return res.status(404).json({ error: "User not found" });
  }

  res.json(user);
});

//
// POST /api/users
// создание пользователя
//
router.post("/", async (req, res) => {
  const db = await connectDB();
  const user = {
    email: req.body.email,
    role: req.body.role || "client",
    createdAt: new Date()
  };

  const result = await db.collection("users").insertOne(user);

  res.status(201).json({
    message: "User created",
    user: {
      _id: result.insertedId,
      ...user
    }
  });
});

module.exports = router;
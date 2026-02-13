const express = require("express");

const usersRoutes = require("./routes/users");
const classesRoutes = require("./routes/classes");
const attendanceRoutes = require("./routes/attendance");

const app = express();

app.use(express.json());

app.get("/", (req, res) => {
  res.json({
    message: "Sport Club REST API",
    endpoints: [
      "/api/users",
      "/api/classes",
      "/api/attendance"
    ]
  });
});

app.use("/api/users", usersRoutes);
app.use("/api/classes", classesRoutes);
app.use("/api/attendance", attendanceRoutes);

module.exports = app;
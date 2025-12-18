const { MongoClient } = require("mongodb");

const url = "mongodb://localhost:27017";
const client = new MongoClient(url);

let db;

async function connectDB(returnClient = false) {
  if (!client.topology?.isConnected()) {
    await client.connect();
    db = client.db("sport_club");
  }

  return returnClient ? client : db;
}

module.exports = connectDB;
# QUICK START

## 1. Start your database-docker
-  ``` $ cd ./docker/ ```
 - ``` $ docker-compose up -d ```

## 2. Initilize database
 - ``` $ cd ./cmd/ ```
 - ``` $ go run main.go -init ```

## 3. Run tests
 - ``` $ cd ./cmd/ ```
 - ``` $ go run main.go -test ```


# Shard(Шардирование)

## Запуск sharded-кластера

```
cd ./docker/
docker-compose -f docker-compose.mongo-sharded.yml up -d
```


# Инициализация Config Server
- ```$ docker exec -it configsvr mongosh --port 27019 ```

```
rs.initiate({
  _id: "configRepl",
  configsvr: true,
  members: [
    { _id: 0, host: "configsvr:27019" }
  ]
}) 
```

# Инициализация shard1
- ```$ docker exec -it shard1 mongosh --port 27018 ```

```
rs.initiate({
  _id: "shardRepl1",
  members: [
    { _id: 0, host: "shard1:27018" }
  ]
})
```

# Инициализация shard2
- ```$ docker exec -it shard2 mongosh --port 27020 ```

```
rs.initiate({
  _id: "shardRepl2",
  members: [
    { _id: 0, host: "shard2:27020" }
  ]
})
```

# Подключение к mongos
- ```$ docker exec -it mongos mongosh --port 27017 ```

# Добавление шардов в кластер
- ```$ sh.addShard("shardRepl1/shard1:27018") ```
- ```$ sh.addShard("shardRepl2/shard2:27020") ```

# Включение шардинга базы данных
- ```$ sh.enableSharding("sport_club") ```
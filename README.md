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

SELECT full_name FROM user_profiles WHERE user_id=1; — выборка с условием.

INSERT INTO users (email, phone) VALUES ('test@mail.com', '+79995553322'); — вставка строки.

UPDATE users SET phone='+79990001122' WHERE id=1; — обновление данных.

DELETE FROM users WHERE id=3; — удаление строки.

CREATE TABLE test (id SERIAL PRIMARY KEY, name TEXT); — создать таблицу.

DROP TABLE test; — удалить таблицу
DROP DATABASE test; — удалить базу
docker exec -it my-postgres psql -U postgres -c "DROP DATABASE sports_club;"

psql commands
\l                          -- список БД
\c mydb                     -- переключиться на mydb
\dt                         -- список таблиц
\d users                    -- структура таблицы users
\x                          -- включить вертикальный вывод
\df                         -- список функций
\du                         -- список ролей
\conninfo                   -- текущее подключение
\q                          -- выход

apt-get update && apt-get install -y nano fish
SELECT COUNT(*) FROM attendance_logs;  -- должно быть 3000000
SELECT COUNT(*) FROM users;            -- 50000
SELECT COUNT(*) FROM payments;         -- 150000

# docker-compose up -d


docker exec -i my-postgres psql -U postgres -d  postgres -f /docker-entrypoint-initdb.d/demo-big-20170815.sql

-U username
-d database

psql commands
\l                          -- список БД
\c mydb                     -- переключиться на mydb
\dt                         -- список таблиц
\d users                    -- структура таблицы users
SELECT * FROM users LIMIT 5;
\x                          -- включить вертикальный вывод
\df                         -- список функций
\du                         -- список ролей
\conninfo                   -- текущее подключение
\q                          -- выход

CREATE DATABASE sports_club ENCODING 'UTF8' LC_COLLATE 'en_US.UTF-8' LC_CTYPE 'en_US.UTF-8' TEMPLATE template0;
\c sports_club
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

apt-get update && apt-get install -y nano fish

psql -U postgres -d sports_club -f ~/init_db.sql
docker exec -it my-postgres psql -U postgres -c "CREATE DATABASE sports_club ENCODING 'UTF8' LC_COLLATE 'en_US.UTF-8' LC_CTYPE 'en_US.UTF-8' TEMPLATE template0";

docker cp init_db.sql my-postgres:/tmp/
docker exec -it my-postgres psql -U postgres -d sports_club -f /tmp/init_db.sql

docker cp generate_3m_bookings.sql my-postgres:/tmp/
docker exec -it my-postgres psql -U postgres -d sports_club -f /tmp/generate_3m_bookings.sql

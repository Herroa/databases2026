-- Создаём репликационного пользователя и даём привилегии
DO
$$
BEGIN
   IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'repluser') THEN
      CREATE ROLE repluser WITH REPLICATION LOGIN PASSWORD 'replpass';
   END IF;
END
$$;
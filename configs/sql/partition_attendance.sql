-- configs/sql/partition_attendance.sql

DROP TABLE IF EXISTS attendance_logs_part CASCADE;

CREATE TABLE attendance_logs_part (
    id SERIAL,
    user_id INT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
) PARTITION BY RANGE (start_time);

-- Создаём партиции: 2023-01 → 2024-06
DO $$
DECLARE
    start_date DATE := '2023-01-01';
    end_date   DATE := '2024-06-01';
    part_name  TEXT;
BEGIN
    WHILE start_date < end_date LOOP
        part_name := 'attendance_logs_part_' || TO_CHAR(start_date, 'YYYY_MM');
        EXECUTE format('
            CREATE TABLE IF NOT EXISTS %I PARTITION OF attendance_logs_part
            FOR VALUES FROM (%L) TO (%L)',
            part_name,
            start_date,
            start_date + INTERVAL '1 month'
        );
        start_date := start_date + INTERVAL '1 month';
    END LOOP;
END $$;

-- ВСТАВЛЯЕМ ВСЕ 3 МЛН СТРОК (без LIMIT!)
INSERT INTO attendance_logs_part (id, user_id, start_time, end_time)
SELECT id, user_id, start_time, end_time
FROM attendance_logs;

-- Индекс (опционально)
CREATE INDEX IF NOT EXISTS idx_attendance_logs_part_start_time ON attendance_logs_part (start_time);

SELECT 'Done. Total rows: ' || COUNT(*) FROM attendance_logs_part;
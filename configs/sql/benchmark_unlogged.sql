-- =============================================
-- ТЕСТ: Logged vs Unlogged таблицы
-- =============================================

-- 1. Создаем тестовые таблицы
CREATE TABLE test_logged (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    event_time TIMESTAMP NOT NULL DEFAULT NOW(),
    payload TEXT
);

CREATE UNLOGGED TABLE test_unlogged (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    event_time TIMESTAMP NOT NULL DEFAULT NOW(),
    payload TEXT
);

-- 2. Вставка пачкой (10 000 записей)
DO $$
DECLARE
    start_t TIMESTAMPTZ;
    end_t   TIMESTAMPTZ;
BEGIN
    RAISE NOTICE '--- INSERT BATCH ---';

    -- Logged
    start_t := clock_timestamp();
    INSERT INTO test_logged (user_id, event_time, payload)
    SELECT g.id, NOW() - (random() * 365 * '1 day'::interval), 'payload_' || g.id
    FROM generate_series(1, 10000) AS g(id);
    end_t := clock_timestamp();
    RAISE NOTICE 'Logged insert batch: % ms', EXTRACT(EPOCH FROM (end_t - start_t)) * 1000;

    -- Unlogged
    start_t := clock_timestamp();
    INSERT INTO test_unlogged (user_id, event_time, payload)
    SELECT g.id, NOW() - (random() * 365 * '1 day'::interval), 'payload_' || g.id
    FROM generate_series(1, 10000) AS g(id);
    end_t := clock_timestamp();
    RAISE NOTICE 'Unlogged insert batch: % ms', EXTRACT(EPOCH FROM (end_t - start_t)) * 1000;

END $$;

-- 3. Чтение (по условию)
DO $$
DECLARE
    start_t TIMESTAMPTZ;
    end_t   TIMESTAMPTZ;
    cnt     BIGINT;
BEGIN
    RAISE NOTICE '--- SELECT BY USER_ID ---';

    -- Logged
    start_t := clock_timestamp();
    SELECT COUNT(*) INTO cnt FROM test_logged WHERE user_id BETWEEN 100 AND 200;
    end_t := clock_timestamp();
    RAISE NOTICE 'Logged select: % ms, rows=%', EXTRACT(EPOCH FROM (end_t - start_t)) * 1000, cnt;

    -- Unlogged
    start_t := clock_timestamp();
    SELECT COUNT(*) INTO cnt FROM test_unlogged WHERE user_id BETWEEN 100 AND 200;
    end_t := clock_timestamp();
    RAISE NOTICE 'Unlogged select: % ms, rows=%', EXTRACT(EPOCH FROM (end_t - start_t)) * 1000, cnt;

END $$;

-- 4. Удаление (по условию)
DO $$
DECLARE
    start_t TIMESTAMPTZ;
    end_t   TIMESTAMPTZ;
BEGIN
    RAISE NOTICE '--- DELETE BY USER_ID ---';

    -- Logged
    start_t := clock_timestamp();
    DELETE FROM test_logged WHERE user_id < 1000;
    end_t := clock_timestamp();
    RAISE NOTICE 'Logged delete: % ms', EXTRACT(EPOCH FROM (end_t - start_t)) * 1000;

    -- Unlogged
    start_t := clock_timestamp();
    DELETE FROM test_unlogged WHERE user_id < 1000;
    end_t := clock_timestamp();
    RAISE NOTICE 'Unlogged delete: % ms', EXTRACT(EPOCH FROM (end_t - start_t)) * 1000;

END $$;

-- 5. Очистка (опционально)
-- DROP TABLE IF EXISTS test_logged, test_unlogged;
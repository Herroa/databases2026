-- Отключаем триггеры и FK на время (опционально, для скорости)
-- Но в нашем случае всё через CASCADE/RESTRICT — лучше оставить.

-- 1. Пользователи (10 000)
INSERT INTO users (email)
SELECT 'user' || g.id || '@example.com'
FROM generate_series(1, 10000) AS g(id);

-- 2. Тренеры (100 первых пользователей)
INSERT INTO coaches (user_id)
SELECT id FROM users WHERE id <= 100;

-- 3. Виды спорта (10)
INSERT INTO sports (name)
VALUES
  ('Football'), ('Basketball'), ('Tennis'), ('Swimming'), ('Yoga'),
  ('Boxing'), ('Cycling'), ('Running'), ('Gym'), ('Martial Arts');

-- 4. Занятия (500)
INSERT INTO classes (sport_id, coach_id)
SELECT
  (random() * 9 + 1)::INT,  -- sport_id от 1 до 10
  (random() * 99 + 1)::INT  -- coach_id от 1 до 100
FROM generate_series(1, 500);

-- 5. Залы (20)
INSERT INTO rooms (capacity)
SELECT (random() * 40 + 10)::INT  -- от 10 до 50 мест
FROM generate_series(1, 20);

-- 6. Расписание (50 000 записей за последние 365 дней)
INSERT INTO schedules (class_id, room_id, start_time, end_time)
SELECT
  (random() * 499 + 1)::INT,
  (random() * 19 + 1)::INT,
  ts.start_time,
  ts.start_time + INTERVAL '1 hour'
FROM (
  SELECT
    NOW() - (random() * 365 * '1 day'::interval) - (random() * 24 * '1 hour'::interval) AS start_time
  FROM generate_series(1, 50000)
) AS ts;

-- 7. Абонементы (5 типов)
INSERT INTO memberships (duration_days, price)
VALUES
  (30, 29.99),
  (90, 79.99),
  (180, 149.99),
  (365, 249.99),
  (7, 9.99);

-- 8. Активные абонементы (по 1 на 80% пользователей)
INSERT INTO user_memberships (user_id, membership_id, started_at, ended_at, is_active)
SELECT
  u.id,
  (random() * 4 + 1)::INT,
  NOW() - (random() * 100)::INT * '1 day'::interval,
  NOW() + (random() * 200 + 30)::INT * '1 day'::interval,
  true
FROM users u
WHERE u.id <= 8000;  -- 80% от 10k

-- 9. Платежи (по 1–3 на пользователя с абонементом)
INSERT INTO payments (user_id, amount, status)
SELECT
  um.user_id,
  m.price,
  'completed'
FROM user_memberships um
JOIN memberships m ON um.membership_id = m.id;

-- 10. Бронирования (пользователь + расписание)
-- Бронируем ~20% расписаний
INSERT INTO bookings (user_id, schedule_id, status)
SELECT DISTINCT
  (random() * 7999 + 1)::INT,  -- user_id от 1 до 8000
  s.id,
  'confirmed'
FROM schedules s
WHERE random() < 0.2
LIMIT 10000;  -- не больше 10k бронирований

-- 11. Посещения (attendance_logs) — 3 000 000+ записей
INSERT INTO attendance_logs (user_id, start_time, end_time)
SELECT
  (random() * 7999 + 1)::INT,
  visit_time,
  visit_time + ((random() * 7200 + 1800) * '1 second'::interval)  -- 30–120 мин
FROM (
  SELECT
    '2023-01-01'::timestamp + (random() * 515 * 86400)::INT * '1 second'::interval AS visit_time
  FROM generate_series(1, 30000)
) AS v;

-- 12. Отзывы
INSERT INTO reviews (user_id, coach_id, class_id, rating)
SELECT
  (random() * 7999 + 1)::INT,
  CASE WHEN r.rand_val < 0.5 THEN (random() * 99 + 1)::INT ELSE NULL::INT END,
  CASE WHEN r.rand_val >= 0.5 THEN (random() * 499 + 1)::INT ELSE NULL::INT END,
  (random() * 4 + 1)::INT
FROM (
  SELECT random() AS rand_val
  FROM generate_series(1, 5000)
) AS r;

-- 13. Промокоды
INSERT INTO promotions (code, discount_percent, valid_from, valid_until, max_uses)
SELECT
  'PROMO' || g.id,
  (random() * 40 + 10)::INT,
  NOW() - '30 days'::interval,
  NOW() + '60 days'::interval,
  (random() * 1000 + 100)::INT
FROM generate_series(1, 10) AS g(id);

-- 14. Использование промокодов (надёжно: только существующие ID)
INSERT INTO promotion_usage (user_id, promotion_id)
SELECT
  (random() * 7999 + 1)::INT,
  p.id
FROM (
  SELECT id FROM promotions
  ORDER BY random()
  LIMIT 2000
) p;

-- 15. Уведомления
INSERT INTO notifications (user_id, is_read)
SELECT
  (random() * 7999 + 1)::INT,
  random() < 0.7
FROM generate_series(1, 15000);

-- 16. Баллы лояльности
INSERT INTO loyalty_points (user_id, points)
SELECT
  u.id,
  (random() * 5000)::INT
FROM users u
WHERE u.id <= 5000;

-- 17. Рефералы
INSERT INTO referrals (referrer_id, referred_id, rewarded)
SELECT
  (random() * 4999 + 1)::INT,
  (random() * 4999 + 5001)::INT,  -- referred_id из другой половины
  random() < 0.9
FROM generate_series(1, 1000);

-- 18. Аудит-логи
INSERT INTO audit_logs (user_id, action, entity_type, entity_id)
SELECT
  (random() * 7999 + 1)::INT,
  CASE (random() * 3)::INT
    WHEN 0 THEN 'booking_created'
    WHEN 1 THEN 'payment_made'
    WHEN 2 THEN 'attendance_logged'
    ELSE 'membership_activated'
  END,
  CASE (random() * 2)::INT
    WHEN 0 THEN 'user'
    WHEN 1 THEN 'booking'
    ELSE 'payment'
  END,
  (random() * 100000)::INT
FROM generate_series(1, 50000);

-- 19. Системные настройки
INSERT INTO system_settings (key, value)
VALUES
  ('club_name', 'FitSport Club'),
  ('max_booking_days_ahead', '14'),
  ('loyalty_points_per_visit', '10');

-- 20. Временные брони (для Redis-интеграции)
INSERT INTO temp_bookings (user_id, schedule_id, expires_at, token)
SELECT
  (random() * 7999 + 1)::INT,
  (random() * 49999 + 1)::INT,
  NOW() + (random() * 30 + 5) * '1 minute'::interval,
  md5(random()::text || clock_timestamp()::text)
FROM generate_series(1, 1000);
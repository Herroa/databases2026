-- Количество посещений по месяцам (последний год)
SELECT date_trunc('month', checked_in) AS mon, COUNT(*) AS visits
FROM attendance_logs
WHERE checked_in > now() - interval '12 months'
GROUP BY mon ORDER BY mon;

-- Среднее количество бронирований на пользователя
SELECT AVG(b_cnt) FROM (
  SELECT user_id, COUNT(*) AS b_cnt FROM bookings GROUP BY user_id
) t;

-- Топ-10 тренеров по числу классов
SELECT c.id, u.email, COUNT(cl.id) AS classes_count
FROM coaches c
JOIN users u ON u.id = c.user_id
LEFT JOIN classes cl ON cl.coach_id = c.id
GROUP BY c.id,u.email
ORDER BY classes_count DESC LIMIT 10;

-- Общая выручка по месяцам
SELECT date_trunc('month', paid_at) AS mon, SUM(amount) AS revenue
FROM payments
GROUP BY mon ORDER BY mon;

-- Средняя длительность класса (аггрегат по классам)
SELECT AVG(duration_minutes) AS avg_duration FROM classes;

-- Рейтинг пользователей по сумме платежей (ранжирование)
SELECT user_id, SUM(amount) AS total_paid,
  RANK() OVER (ORDER BY SUM(amount) DESC) AS rk
FROM payments
GROUP BY user_id
ORDER BY total_paid DESC LIMIT 20;

-- Скользящая сумма посещений по дате (7-дневное окно)
SELECT checked_in::date AS d,
  COUNT(*) OVER (ORDER BY checked_in::date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) AS rolling_7d
FROM attendance_logs
WHERE checked_in > now() - interval '30 days'
GROUP BY checked_in::date
ORDER BY d;

-- Оконная функция: предыдущий визит пользователя
SELECT id, user_id, checked_in,
  LAG(checked_in) OVER (PARTITION BY user_id ORDER BY checked_in) AS prev_visit
FROM attendance_logs
WHERE user_id < 200
ORDER BY user_id, checked_in
LIMIT 100;

-- Оценка тренеров: средний рейтинг + отклонение по тренеру
SELECT coach_id, AVG(rating) OVER (PARTITION BY coach_id) AS avg_rating, COUNT(*) OVER (PARTITION BY coach_id) AS cnt
FROM reviews
ORDER BY avg_rating DESC
LIMIT 20;

-- Кумулятивная выручка
SELECT paid_at::date AS day, SUM(amount) AS daily, SUM(SUM(amount)) OVER (ORDER BY paid_at::date) AS cumulative
FROM payments
GROUP BY paid_at::date
ORDER BY day;

-- J2-1: bookings JOIN users (2 tables)
SELECT b.id, u.email, s.start_time, b.status
FROM bookings b
JOIN users u ON u.id = b.user_id
JOIN schedules s ON s.id = b.schedule_id
LIMIT 50;

-- J2-2: reviews JOIN coaches (2 tables)
SELECT r.id, u.email AS reviewer, c.id AS coach_id, r.rating, r.comment
FROM reviews r
JOIN users u ON u.id = r.user_id
JOIN coaches c ON c.id = r.coach_id
LIMIT 50;

-- J3-1: bookings -> users -> schedules
SELECT b.id, u.email, s.start_time, cl.name AS class_name
FROM bookings b
JOIN users u ON u.id = b.user_id
JOIN schedules s ON s.id = b.schedule_id
JOIN classes cl ON cl.id = s.class_id
LIMIT 50;

-- J3-2: attendance_logs -> users -> schedules
SELECT a.id, u.email, s.start_time, cl.name
FROM attendance_logs a
JOIN users u ON u.id = a.user_id
JOIN schedules s ON s.id = a.schedule_id
JOIN classes cl ON cl.id = s.class_id
LIMIT 50;

-- J3-3: payments -> users -> memberships (via user_memberships)
SELECT p.id, u.email, um.membership_id, m.name
FROM payments p
JOIN users u ON u.id = p.user_id
LEFT JOIN user_memberships um ON um.user_id = u.id
LEFT JOIN memberships m ON m.id = um.membership_id
LIMIT 50;

-- J3-4: reviews -> users -> coaches (and coaches -> users)
SELECT r.id, u.email AS reviewer, c.user_id AS coach_user_id, rc.email AS coach_email, r.rating
FROM reviews r
JOIN users u ON u.id = r.user_id
JOIN coaches c ON c.id = r.coach_id
LEFT JOIN users rc ON rc.id = c.user_id
LIMIT 50;

-- J4: bookings -> users -> schedules -> classes -> rooms   (actually 5 tables, but we'll show a 4-table)
SELECT b.id, u.email, s.start_time, cl.name AS class_name, r.name AS room_name
FROM bookings b
JOIN users u ON u.id = b.user_id
JOIN schedules s ON s.id = b.schedule_id
JOIN classes cl ON cl.id = s.class_id
JOIN rooms r ON r.id = s.room_id
LIMIT 50;

-- J5: payments -> users -> user_memberships -> memberships -> (maybe referrals) to get 5 tables
SELECT p.id, u.email, um.started_at, m.name AS membership_name, ref.referrer_id
FROM payments p
JOIN users u ON u.id = p.user_id
LEFT JOIN user_memberships um ON um.user_id = u.id
LEFT JOIN memberships m ON m.id = um.membership_id
LEFT JOIN referrals ref ON ref.referred_id = u.id
LIMIT 50;

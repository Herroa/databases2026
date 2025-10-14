-- =============================================
-- БИЗНЕС-ЗАПРОСЫ
-- =============================================

-- 1. Топ-5 тренеров по количеству посещений за последний месяц
SELECT c.user_id AS coach_id, COUNT(*) AS visits
FROM attendance_logs a
JOIN bookings b ON a.user_id = b.user_id
JOIN schedules s ON b.schedule_id = s.id
JOIN classes cl ON s.class_id = cl.id
JOIN coaches c ON cl.coach_id = c.user_id
WHERE a.start_time >= NOW() - INTERVAL '1 month'
GROUP BY c.user_id
ORDER BY visits DESC
LIMIT 5;

-- 2. Активные абонементы пользователей
SELECT u.email, m.duration_days, um.started_at, um.ended_at
FROM users u
JOIN user_memberships um ON u.id = um.user_id
JOIN memberships m ON um.membership_id = m.id
WHERE um.is_active = true
  AND um.ended_at > NOW();

-- 3. Загруженность залов на ближайшие 7 дней
SELECT r.id, r.capacity, COUNT(b.id) AS confirmed_bookings
FROM rooms r
LEFT JOIN schedules s ON r.id = s.room_id
LEFT JOIN bookings b ON s.id = b.schedule_id AND b.status = 'confirmed'
WHERE s.start_time >= NOW() AND s.start_time < NOW() + INTERVAL '7 days'
GROUP BY r.id, r.capacity
ORDER BY confirmed_bookings DESC;

-- 4. Пользователи с наибольшим количеством отзывов
SELECT u.email, COUNT(r.id) AS review_count
FROM users u
JOIN reviews r ON u.id = r.user_id
GROUP BY u.email
ORDER BY review_count DESC
LIMIT 10;
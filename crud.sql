-- psql -d sports_club -f crud.sql
SET client_min_messages = WARNING;

INSERT INTO users (email, phone) VALUES ('sample1@example.com','+70000000001');
SELECT * FROM users LIMIT 5;
SELECT * FROM users WHERE id = 1;
UPDATE users SET phone = '+71111111111', updated_at = now() WHERE id = 1;
DELETE FROM users WHERE id = 999999;

INSERT INTO user_profiles (user_id, full_name, birth_date, gender, address) VALUES (1,'Иван Тест','1990-01-01','male','г.Москва');
SELECT * FROM user_profiles LIMIT 5;
SELECT * FROM user_profiles WHERE id = 1;
UPDATE user_profiles SET full_name='Иван Нов', updated_at=now() WHERE id = 1;
DELETE FROM user_profiles WHERE id = 999999;

INSERT INTO coaches (user_id, specialization, hire_date) VALUES (3,'Йога','2021-01-01');
SELECT * FROM coaches LIMIT 5;
SELECT * FROM coaches WHERE id = 1;
UPDATE coaches SET specialization='Аэробика' WHERE id = 1;
DELETE FROM coaches WHERE id = 999999;

INSERT INTO sports (name, description) VALUES ('Пилатес','описание');
SELECT * FROM sports LIMIT 5;
SELECT * FROM sports WHERE id = 1;
UPDATE sports SET description='новое описание' WHERE id = 1;
DELETE FROM sports WHERE id = 999999;

INSERT INTO classes (name, sport_id, coach_id, description, duration_minutes, max_participants)
VALUES ('Утренняя тренировка',1,1,'Описание',60,20);
SELECT * FROM classes LIMIT 5;
SELECT * FROM classes WHERE id = 1;
UPDATE classes SET max_participants=25 WHERE id = 1;
DELETE FROM classes WHERE id = 999999;

INSERT INTO rooms (name, capacity, location) VALUES ('Зал А',20,'Первый этаж');
SELECT * FROM rooms LIMIT 5;
SELECT * FROM rooms WHERE id = 1;
UPDATE rooms SET capacity=30 WHERE id = 1;
DELETE FROM rooms WHERE id = 999999;

INSERT INTO schedules (class_id, room_id, start_time, end_time) VALUES (1,1, now()+interval '1 day', now()+interval '1 day' + interval '1 hour');
SELECT * FROM schedules LIMIT 5;
SELECT * FROM schedules WHERE id = 1;
UPDATE schedules SET is_cancelled = TRUE WHERE id = 1;
DELETE FROM schedules WHERE id = 999999;

INSERT INTO bookings (user_id, schedule_id, status) VALUES (1,1,'confirmed');
SELECT * FROM bookings LIMIT 5;
SELECT * FROM bookings WHERE id = 1;
UPDATE bookings SET status='cancelled' WHERE id = 1;
DELETE FROM bookings WHERE id = 999999;

INSERT INTO memberships (name, duration_days, price, description) VALUES ('Пробный',7,500.00,'тест');
SELECT * FROM memberships LIMIT 5;
SELECT * FROM memberships WHERE id = 1;
UPDATE memberships SET price = 600 WHERE id = 1;
DELETE FROM memberships WHERE id = 999999;

INSERT INTO user_memberships (user_id, membership_id, started_at, ended_at) VALUES (1,1, now()::date, now()::date + interval '30 days');
SELECT * FROM user_memberships LIMIT 5;
SELECT * FROM user_memberships WHERE id = 1;
UPDATE user_memberships SET is_active = FALSE WHERE id = 1;
DELETE FROM user_memberships WHERE id = 999999;

INSERT INTO payments (user_id, amount, currency, payment_method, status) VALUES (1,100.00,'RUB','card','completed');
SELECT * FROM payments LIMIT 5;
SELECT * FROM payments WHERE id = 1;
UPDATE payments SET status='refunded' WHERE id = 1;
DELETE FROM payments WHERE id = 999999;

INSERT INTO attendance_logs (user_id, schedule_id, checked_in, checked_out) VALUES (1,1, now(), now() + interval '60 minutes');
SELECT * FROM attendance_logs LIMIT 5;
SELECT * FROM attendance_logs WHERE id = 1;
UPDATE attendance_logs SET checked_out = now() WHERE id = 1;
DELETE FROM attendance_logs WHERE id = 999999;

INSERT INTO reviews (user_id, coach_id, class_id, rating, comment) VALUES (1,1,1,5,'Отлично');
SELECT * FROM reviews LIMIT 5;
SELECT * FROM reviews WHERE id = 1;
UPDATE reviews SET rating=4 WHERE id = 1;
DELETE FROM reviews WHERE id = 999999;

INSERT INTO promotions (code, discount_percent, valid_from, valid_until, max_uses) VALUES ('TEST10',10, now()::date, now()::date + interval '30 days', 100);
SELECT * FROM promotions LIMIT 5;
SELECT * FROM promotions WHERE id = 1;
UPDATE promotions SET used_count = used_count + 1 WHERE id = 1;
DELETE FROM promotions WHERE id = 999999;

INSERT INTO promotion_usage (user_id, promotion_id) VALUES (1,1);
SELECT * FROM promotion_usage LIMIT 5;
SELECT * FROM promotion_usage WHERE id = 1;
UPDATE promotion_usage SET used_at = now() WHERE id = 1;
DELETE FROM promotion_usage WHERE id = 999999;

INSERT INTO notifications (user_id, title, message) VALUES (1,'Привет','Сообщение');
SELECT * FROM notifications LIMIT 5;
SELECT * FROM notifications WHERE id = 1;
UPDATE notifications SET is_read = TRUE WHERE id = 1;
DELETE FROM notifications WHERE id = 999999;

INSERT INTO loyalty_points (user_id, points) VALUES (1,100);
SELECT * FROM loyalty_points LIMIT 5;
SELECT * FROM loyalty_points WHERE id = 1;
UPDATE loyalty_points SET points = points + 10 WHERE id = 1;
DELETE FROM loyalty_points WHERE id = 999999;

INSERT INTO referrals (referrer_id, referred_id, rewarded) VALUES (1,3, FALSE);
SELECT * FROM referrals LIMIT 5;
SELECT * FROM referrals WHERE id = 1;
UPDATE referrals SET rewarded = TRUE WHERE id = 1;
DELETE FROM referrals WHERE id = 999999;

INSERT INTO audit_logs (user_id, action, table_name, record_id, details) VALUES (1,'INSERT','users',1,'{"note":"test"}');
SELECT * FROM audit_logs LIMIT 5;
SELECT * FROM audit_logs WHERE id = 1;
UPDATE audit_logs SET details = details || '{"updated":true}' WHERE id = 1;
DELETE FROM audit_logs WHERE id = 999999;

INSERT INTO system_settings (setting_key, setting_value, description) VALUES ('timezone','Europe/Moscow','Системная таймзона');
SELECT * FROM system_settings LIMIT 5;
SELECT * FROM system_settings WHERE id = 1;
UPDATE system_settings SET setting_value = 'UTC' WHERE id = 1;
DELETE FROM system_settings WHERE id = 999999;

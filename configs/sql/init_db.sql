-- 1. Пользователи (только для связи и аутентификации)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL
);

-- 2. Тренеры (роль через наличие записи)
CREATE TABLE coaches (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE
);

-- 3. Виды спорта
CREATE TABLE sports (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- 4. Занятия
CREATE TABLE classes (
    id SERIAL PRIMARY KEY,
    sport_id INT NOT NULL REFERENCES sports(id) ON DELETE RESTRICT,
    coach_id INT NOT NULL REFERENCES coaches(user_id) ON DELETE RESTRICT
);

-- 5. Залы
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    capacity INT NOT NULL CHECK (capacity > 0)
);

-- 6. Расписание
CREATE TABLE schedules (
    id SERIAL PRIMARY KEY,
    class_id INT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    room_id INT NOT NULL REFERENCES rooms(id) ON DELETE RESTRICT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    CHECK (end_time > start_time)
);

-- 7. Бронирования
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id INT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled')),
    UNIQUE (user_id, schedule_id)
);

-- 8. Абонементы
CREATE TABLE memberships (
    id SERIAL PRIMARY KEY,
    duration_days INT NOT NULL CHECK (duration_days > 0),
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0)
);

-- 9. Активные абонементы
CREATE TABLE user_memberships (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    membership_id INT NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
    started_at DATE NOT NULL,
    ended_at DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

-- 10. Платежи
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL CHECK (amount >= 0),
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('completed', 'failed'))
);

-- 11. Посещения
CREATE TABLE attendance_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
);

-- 12. Отзывы
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coach_id INT REFERENCES coaches(user_id) ON DELETE SET NULL,
    class_id INT REFERENCES classes(id) ON DELETE SET NULL,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    CHECK (coach_id IS NOT NULL OR class_id IS NOT NULL)
);

-- 13. Промокоды
CREATE TABLE promotions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_percent INT NOT NULL CHECK (discount_percent BETWEEN 1 AND 100),
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    max_uses INT,
    used_count INT DEFAULT 0
);

-- 14. Использование промокодов
CREATE TABLE promotion_usage (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    promotion_id INT NOT NULL REFERENCES promotions(id) ON DELETE CASCADE
);

-- 15. Уведомления (минимум)
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_read BOOLEAN DEFAULT FALSE
);

-- 16. Баллы лояльности
CREATE TABLE loyalty_points (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    points INT DEFAULT 0 CHECK (points >= 0)
);

-- 17. Рефералы
CREATE TABLE referrals (
    id SERIAL PRIMARY KEY,
    referrer_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    referred_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rewarded BOOLEAN DEFAULT FALSE,
    CHECK (referrer_id != referred_id)
);

-- 18. Аудит-логи (для событий, которые триггерят синхронизацию)
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INT,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),  -- 'user', 'booking', 'payment'
    entity_id INT,
    performed_at TIMESTAMP DEFAULT NOW()
);

-- 19. Системные настройки
CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT
);

-- 20. Временные брони (для интеграции с Redis)
CREATE TABLE temp_bookings (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    schedule_id INT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    token VARCHAR(64) UNIQUE NOT NULL  -- для идентификации в Redis
);
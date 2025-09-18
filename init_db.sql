-- Создаём базу, если её ещё нет (выполнять от пользователя postgres)
-- CREATE DATABASE sports_club ENCODING 'UTF8' LC_COLLATE 'en_US.UTF-8' LC_CTYPE 'en_US.UTF-8' TEMPLATE template0;
-- \c sports_club

-- 1. Пользователи
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. Профили пользователей
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    full_name VARCHAR(200),
    birth_date DATE,
    gender VARCHAR(10),
    address TEXT,
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 3. Тренеры
CREATE TABLE coaches (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    specialization TEXT,
    hire_date DATE,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 4. Виды спорта
CREATE TABLE sports (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. Занятия
CREATE TABLE classes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    sport_id INT NOT NULL,
    coach_id INT NOT NULL,
    description TEXT,
    duration_minutes INT,
    max_participants INT,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE RESTRICT,
    FOREIGN KEY (coach_id) REFERENCES coaches(id) ON DELETE RESTRICT
);

-- 6. Залы
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    capacity INT CHECK (capacity > 0),
    location TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 7. Расписание
CREATE TABLE schedules (
    id SERIAL PRIMARY KEY,
    class_id INT NOT NULL,
    room_id INT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    is_cancelled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE RESTRICT
);

-- 8. Бронирования
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    schedule_id INT NOT NULL,
    booked_at TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled', 'pending')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE
);

-- 9. Абонементы
CREATE TABLE memberships (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    duration_days INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 10. Привязка абонементов к пользователям
CREATE TABLE user_memberships (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    membership_id INT NOT NULL,
    started_at DATE NOT NULL,
    ended_at DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (membership_id) REFERENCES memberships(id) ON DELETE CASCADE
);

-- 11. Платежи
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL CHECK (amount >= 0),
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50),
    paid_at TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('completed', 'failed', 'pending', 'refunded')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 12. Логи посещений
CREATE TABLE attendance_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    schedule_id INT NOT NULL,
    checked_in TIMESTAMP DEFAULT NOW(),
    checked_out TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE
);

-- 13. Отзывы
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    coach_id INT,
    class_id INT,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (coach_id) REFERENCES coaches(id) ON DELETE SET NULL,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE SET NULL
);

-- 14. Промокоды
CREATE TABLE promotions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_percent INT CHECK (discount_percent BETWEEN 1 AND 100),
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    max_uses INT,
    used_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 15. Использование промокодов
CREATE TABLE promotion_usage (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    promotion_id INT NOT NULL,
    used_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (promotion_id) REFERENCES promotions(id) ON DELETE CASCADE
);

-- 16. Уведомления
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(200),
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    sent_at TIMESTAMP DEFAULT NOW(),
    read_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 17. Баллы лояльности
CREATE TABLE loyalty_points (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    points INT DEFAULT 0 CHECK (points >= 0),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 18. Рефералы
CREATE TABLE referrals (
    id SERIAL PRIMARY KEY,
    referrer_id INT NOT NULL,  -- кто пригласил
    referred_id INT NOT NULL,  -- кого пригласили
    rewarded BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (referrer_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (referred_id) REFERENCES users(id) ON DELETE CASCADE,
    CHECK (referrer_id != referred_id)
);

-- 19. Аудит-логи
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INT,
    action VARCHAR(100) NOT NULL,
    table_name VARCHAR(100),
    record_id INT,
    performed_at TIMESTAMP DEFAULT NOW(),
    details JSONB
);

-- 20. Системные настройки
CREATE TABLE system_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value TEXT,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW()
);
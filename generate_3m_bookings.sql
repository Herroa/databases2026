INSERT INTO attendance_logs (user_id, schedule_id, checked_in, checked_out)
SELECT
    (random() * 1000)::INT + 1,
    (random() * 500)::INT + 1,
    NOW() - (random() * interval '365 days'),
    NOW() - (random() * interval '365 days')
FROM generate_series(1, 3000000);

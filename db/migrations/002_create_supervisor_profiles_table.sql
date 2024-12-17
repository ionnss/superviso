CREATE TABLE IF NOT EXISTS supervisor_profiles (
    user_id INT PRIMARY KEY REFERENCES users(id),
    session_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 

CREATE TABLE IF NOT EXISTS supervisor_weekly_hours (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    weekday INT CHECK (weekday BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    UNIQUE (supervisor_id, weekday)
); 
CREATE TABLE IF NOT EXISTS supervisor_weekly_hours (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    weekday INT CHECK (weekday BETWEEN 1 AND 7),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    UNIQUE (supervisor_id, weekday)
);

CREATE TABLE IF NOT EXISTS supervisor_availability_periods (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_supervisor_period UNIQUE (supervisor_id)
);
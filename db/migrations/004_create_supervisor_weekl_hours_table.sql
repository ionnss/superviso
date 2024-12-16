CREATE TABLE supervisor_weekly_hours (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    weekday INT CHECK (weekday BETWEEN 1 AND 7),
    start_time TIME,
    end_time TIME,
    UNIQUE (supervisor_id, weekday)
);

CREATE TABLE supervisor_availability_periods (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (supervisor_id, start_date, end_date)
);

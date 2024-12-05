CREATE TABLE IF NOT EXISTS supervisor_schedules (
    id SERIAL PRIMARY KEY,
    supervisor_id INTEGER REFERENCES users(id),
    day_of_week INTEGER NOT NULL, -- 1 = Domingo, 2 = Segunda, etc.
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    session_duration INTEGER NOT NULL, -- em minutos
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_day CHECK (day_of_week BETWEEN 1 AND 7),
    CONSTRAINT valid_duration CHECK (session_duration > 0),
    CONSTRAINT valid_price CHECK (price >= 0)
);

CREATE INDEX idx_supervisor_schedules_supervisor ON supervisor_schedules(supervisor_id); 
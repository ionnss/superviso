CREATE TABLE IF NOT EXISTS supervisor_weekly_hours (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    weekday INT CHECK (weekday BETWEEN 1 AND 7),
    start_time TIME,
    end_time TIME,
    UNIQUE (supervisor_id, weekday)
);

CREATE TABLE IF NOT EXISTS supervisor_availability_periods (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (supervisor_id, start_date, end_date)
);

-- Criar índices apenas se não existirem
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_weekly_hours_supervisor') THEN
        CREATE INDEX idx_weekly_hours_supervisor ON supervisor_weekly_hours(supervisor_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_availability_periods_supervisor') THEN
        CREATE INDEX idx_availability_periods_supervisor ON supervisor_availability_periods(supervisor_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_availability_periods_dates') THEN
        CREATE INDEX idx_availability_periods_dates ON supervisor_availability_periods(start_date, end_date);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS available_slots (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    slot_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status VARCHAR(20) DEFAULT 'available',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (supervisor_id, slot_date, start_time),
    CHECK (slot_date >= CURRENT_DATE)
);

CREATE TABLE IF NOT EXISTS appointments (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    supervisee_id INT REFERENCES users(id),
    slot_id INT REFERENCES available_slots(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    cancellation_reason TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_slot_booking UNIQUE (slot_id)
);

-- Índices para melhor performance
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_slots_supervisor') THEN
        CREATE INDEX idx_slots_supervisor ON available_slots(supervisor_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_slots_date') THEN
        CREATE INDEX idx_slots_date ON available_slots(slot_date);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_appointments_supervisor') THEN
        CREATE INDEX idx_appointments_supervisor ON appointments(supervisor_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_appointments_supervisee') THEN
        CREATE INDEX idx_appointments_supervisee ON appointments(supervisee_id);
    END IF;
END $$; 
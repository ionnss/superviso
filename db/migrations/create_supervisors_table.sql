CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    qualifications TEXT NOT NULL,
    price_per_session NUMERIC(10, 2),
    availability TEXT,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    cpf VARCHAR(11) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    crp VARCHAR(20) UNIQUE NOT NULL,
    theory_approach VARCHAR(100) NOT NULL,
    qualifications TEXT NOT NULL,
    user_role VARCHAR(20) NOT NULL,
    price_per_session NUMERIC(10, 2),
    sessions_availability TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
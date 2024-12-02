-- api/db/create_users_table.sql

-- tabela usuários supervisores
CREATE TABLE IF NOT EXISTS supervisor (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    cpf VARCHAR(11) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    crp VARCHAR(20) UNIQUE NOT NULL,
    theory_approach VARCHAR(100) NOT NULL,
    qualifications TEXT NOT NULL,
    user_role VARCHAR(20) NOT NULL DEFAULT 'supervisor',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);

-- tabela usuários supervisionados
CREATE TABLE IF NOT EXISTS supervisionated (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    cpf VARCHAR(11) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    crp VARCHAR(20) UNIQUE NOT NULL,
    theory_approach VARCHAR(100) NOT NULL,
    qualifications TEXT NOT NULL,
    user_role VARCHAR(20) NOT NULL DEFAULT 'supervisionado',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);

-- tabela de disponibilidade de supervisores
CREATE TABLE IF NOT EXISTS supervisor_availability (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    availability_day VARCHAR(20), -- Dias da semana (segunda, terça, etc.)
    availability_time TIME,       -- Horário correspondente
    price_per_session NUMERIC(10, 2) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES supervisor(id) ON DELETE CASCADE
);

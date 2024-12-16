CREATE TABLE IF NOT EXISTS supervisor_profiles (
    user_id INT PRIMARY KEY REFERENCES users(id),
    user_crp VARCHAR(20) REFERENCES users(crp),
    session_price DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 
CREATE TABLE IF NOT EXISTS supervisor_profiles (
    user_id INT PRIMARY KEY REFERENCES users(id),
    session_price DECIMAL(10,2),
    available_days VARCHAR(100),  -- Ex: "1,2,3,4,5" (seg a sex)
    start_time TIME,
    end_time TIME,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 
CREATE TABLE IF NOT EXISTS otps (
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    code CHAR(6) NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, code)
);
CREATE TYPE user_provider AS ENUM('yandex', 'google', 'email');

CREATE TABLE users (
    user_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    provider user_provider NOT NULL,
    is_email_verified BOOLEAN DEFAULT FALSE,
    avatar_url TEXT,
    full_name TEXT,
    created_at TIMESTAMP DEFAULT now()
);
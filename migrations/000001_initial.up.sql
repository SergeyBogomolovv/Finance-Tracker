CREATE TYPE currency AS ENUM('RUB', 'USD');
CREATE TYPE frequency AS ENUM('daily', 'weekly', 'monthly', 'quarterly', 'half_yearly', 'yearly', 'once');
CREATE TYPE sub_status AS ENUM('active', 'cancelled', 'trial', 'expired');
CREATE TYPE notification_status AS ENUM('sent', 'failed');
CREATE TYPE notification_type AS ENUM('payment_reminder', 'otp', '2fa', 'email_change', 'weekly_summary');
CREATE TYPE task_status AS ENUM('pending', 'completed', 'failed');

CREATE TABLE users (
    user_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash BYTEA, -- если гугл - будет null
    google_id TEXT UNIQUE, -- если только пароль - будет null
    is_email_verified BOOLEAN DEFAULT FALSE,
    avatar_url TEXT,
    full_name TEXT,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE email_otps (
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    code CHAR(6) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    PRIMARY KEY (user_id, code)
);

CREATE TABLE subscriptions (
    sub_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    notes TEXT,
    cost NUMERIC(10, 2) NOT NULL,
    currency currency DEFAULT 'RUB',
    frequency frequency NOT NULL,
    notifications_enabled BOOLEAN DEFAULT TRUE,
    auto_payment BOOLEAN DEFAULT FALSE,
    status sub_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    start_date DATE NOT NULL,
    next_payment_date DATE NOT NULL,
    last_payment_date DATE,
    trial_end_date DATE
);

CREATE TABLE expenses (
    expense_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    subscription_id INT REFERENCES subscriptions(sub_id) ON DELETE SET NULL,
    amount NUMERIC(10, 2) NOT NULL,
    currency currency DEFAULT 'RUB',
    is_auto BOOLEAN DEFAULT FALSE,
    paid_at DATE NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE notifications (
    notification_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    status notification_status NOT NULL,
    sent_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE scheduled_tasks (
    task_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT NOT NULL,
    task_type TEXT NOT NULL, -- например: 'payment_reminder', 'autopayment', 'weekly_summary'
    payload JSONB, -- данные, которые понадобятся исполнителю задачи
    run_at TIMESTAMP NOT NULL, -- когда нужно запустить
    status task_status NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    created_at TIMESTAMP DEFAULT now()
);

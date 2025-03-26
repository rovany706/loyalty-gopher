CREATE TYPE e_accrual_status AS ENUM (
    'REGISTERED',
    'INVALID',
    'PROCESSED',
    'PROCESSING'
);

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username TEXT UNIQUE NOT NULL,
    pw_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    order_num TEXT UNIQUE NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    accrual_status e_accrual_status NOT NULL,
    accrual NUMERIC(12,2),
    user_id INT REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS point_accounts (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    balance NUMERIC(12,2) NOT NULL,
    user_id INT REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS withdrawal_history (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    order_num TEXT NOT NULL,
    amount NUMERIC(12,2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    point_account_id INT REFERENCES point_accounts(id)
);

CREATE TYPE transaction_type_enum AS ENUM ('bet', 'win');

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    transaction_type transaction_type_enum NOT NULL,
    amount NUMERIC(10, 2) NOT NULL CHECK (amount >= 0),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
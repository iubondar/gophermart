-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS orders (
    uploaded_at TIMESTAMP DEFAULT NOW(),
    id SERIAL PRIMARY KEY,
    user_id uuid,
    order_number VARCHAR UNIQUE,
    order_status VARCHAR(20),
    accrual float DEFAULT 0
);

CREATE INDEX IF NOT EXISTS order_user_id_idx ON orders (user_id);

CREATE INDEX IF NOT EXISTS order_status_idx ON orders (order_status);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP INDEX IF EXISTS order_status_idx;

DROP INDEX IF EXISTS order_user_id_idx;

DROP TABLE IF EXISTS orders;
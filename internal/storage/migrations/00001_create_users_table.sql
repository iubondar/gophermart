-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS users (
    created_at TIMESTAMP DEFAULT NOW(),
    user_id uuid PRIMARY KEY,
	user_name VARCHAR(100) UNIQUE NOT NULL,
	password_hash VARCHAR(100) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS user_id_idx ON users (user_id);

CREATE UNIQUE INDEX IF NOT EXISTS user_name_idx ON users (user_id);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP INDEX IF EXISTS user_name_idx;

DROP INDEX IF EXISTS user_id_idx;

DROP TABLE IF EXISTS users;
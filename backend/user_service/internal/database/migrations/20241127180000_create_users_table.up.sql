-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  login TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  registration_date TIMESTAMP NOT NULL,
  role VARCHAR NOT NULL,
  name TEXT NOT NULL,
  height INTEGER NOT NULL,
  weight INTEGER NOT NULL,
  age INTEGER NOT NULL,
  password TEXT NOT NULL
);

CREATE INDEX idx_users_login ON users(login);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd

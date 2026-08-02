-- +goose Up
CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY,
                                     login VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
    );

-- +goose Down
DROP TABLE IF EXISTS users;

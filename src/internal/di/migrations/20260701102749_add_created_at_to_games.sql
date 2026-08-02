-- +goose Up
-- +goose StatementBegin
ALTER TABLE games ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games DROP COLUMN IF EXISTS created_at;
-- +goose StatementEnd
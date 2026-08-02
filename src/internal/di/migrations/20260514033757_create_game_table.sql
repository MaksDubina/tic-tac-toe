-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS games (
                                     id UUID PRIMARY KEY,
                                     board JSONB NOT NULL,
                                     player_x_id UUID,
                                     player_o_id UUID,
                                     status INT NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS games;
-- +goose StatementEnd

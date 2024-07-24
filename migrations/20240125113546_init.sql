-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS chunks (
    block BIGSERIAL,
    author TEXT NOT NULL,
    chunk NUMERIC(78, 0),
    inserted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_block ON chunks(block);
CREATE INDEX IF NOT EXISTS idx_chunk ON chunks(chunk);
CREATE INDEX IF NOT EXISTS idx_author ON chunks(author);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chunks;
-- +goose StatementEnd

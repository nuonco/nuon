-- +goose Up
CREATE TABLE artifacts (
    id text PRIMARY KEY,
    created_at timestamptz NULL,
    updated_at timestamptz NULL,
    deleted_at timestamptz NULL,
    created_by_id text NULL
);

-- +goose Down
DROP TABLE IF EXISTS artifacts;

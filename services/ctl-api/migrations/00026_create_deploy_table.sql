-- +goose Up
-- +goose StatementBegin

CREATE TABLE deploys (
    id text PRIMARY KEY,
    created_at timestamptz NULL,
    updated_at timestamptz NULL,
    deleted_at timestamptz NULL,
    build_id text REFERENCES builds(id),
    install_id text REFERENCES installs(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS deploys;
-- +goose StatementEnd

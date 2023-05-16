-- +goose Up
-- +goose StatementBegin
-- apps definition

CREATE TABLE builds (
    id text PRIMARY KEY,
    created_at timestamptz NULL,
    updated_at timestamptz NULL,
    deleted_at timestamptz NULL,
    created_by_id text NULL,
    component_id text REFERENCES components(id) NOT NULL,
    git_ref text NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop table

DROP TABLE builds;
-- +goose StatementEnd

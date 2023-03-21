-- +goose Up
-- +goose StatementBegin
ALTER TABLE components ADD COLUMN config JSONB NOT NULL DEFAULT '{}'::jsonb;
-- +goose StatementEnd

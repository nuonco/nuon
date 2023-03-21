-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE apps DROP COLUMN IF EXISTS slug;
ALTER TABLE orgs DROP COLUMN IF EXISTS workflow_id;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE orgs DROP COLUMN IF EXISTS slug;
DROP INDEX IF EXISTS idx_orgs_name;
DROP INDEX IF EXISTS idx_orgs_slug;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

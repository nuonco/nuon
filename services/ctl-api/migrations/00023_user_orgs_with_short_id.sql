-- +goose Up
-- +goose StatementBegin

-- alter user_orgs.id to text + remove default value
ALTER TABLE user_orgs ALTER COLUMN id TYPE TEXT;
ALTER TABLE user_orgs ALTER COLUMN id DROP DEFAULT;

-- +goose StatementEnd

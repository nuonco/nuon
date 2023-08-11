-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- alter org_id to uuid
ALTER TABLE apps ALTER COLUMN org_id TYPE uuid USING org_id::uuid;

-- add forgeign key contraint to orgs table
ALTER TABLE apps ADD CONSTRAINT fk_orgs_apps FOREIGN KEY (org_id) REFERENCES orgs(id);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

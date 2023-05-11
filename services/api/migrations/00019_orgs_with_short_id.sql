-- +goose Up
-- +goose StatementBegin

-- drop constraints that reference orgs.id
ALTER TABLE apps DROP CONSTRAINT IF EXISTS fk_orgs_apps;
ALTER TABLE user_orgs DROP CONSTRAINT IF EXISTS fk_user_orgs_org;

-- alter all references to orgs.id to TEXT
ALTER TABLE orgs ALTER COLUMN id TYPE TEXT;
ALTER TABLE orgs ALTER COLUMN id DROP DEFAULT;
ALTER TABLE apps ALTER COLUMN org_id TYPE TEXT;
ALTER TABLE user_orgs ALTER COLUMN org_id TYPE TEXT;

-- re-create constraints that reference orgs.id
ALTER TABLE apps ADD CONSTRAINT fk_orgs_apps FOREIGN KEY (org_id) REFERENCES orgs(id);
ALTER TABLE user_orgs ADD CONSTRAINT fk_user_orgs_org FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE;

-- +goose StatementEnd

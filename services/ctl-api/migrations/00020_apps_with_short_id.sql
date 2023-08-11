-- +goose Up
-- +goose StatementBegin

-- drop constraints that reference apps.id
ALTER TABLE components DROP CONSTRAINT IF EXISTS fk_apps_components;
ALTER TABLE installs DROP CONSTRAINT IF EXISTS fk_apps_installs;

-- alter all references to orgs.id to TEXT
ALTER TABLE apps ALTER COLUMN id TYPE TEXT;
ALTER TABLE apps ALTER COLUMN id DROP DEFAULT;
ALTER TABLE components ALTER COLUMN app_id TYPE TEXT;
ALTER TABLE installs ALTER COLUMN app_id TYPE TEXT;

-- re-create constraints that reference apps.id
ALTER TABLE components ADD CONSTRAINT fk_apps_components FOREIGN KEY (app_id) REFERENCES apps(id);
ALTER TABLE installs ADD CONSTRAINT fk_apps_installs FOREIGN KEY (app_id) REFERENCES apps(id);

-- +goose StatementEnd
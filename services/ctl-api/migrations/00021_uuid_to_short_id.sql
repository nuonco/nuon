-- +goose Up
-- +goose StatementBegin

-- drop constraints
ALTER TABLE deployments DROP CONSTRAINT IF EXISTS fk_components_deployments;
ALTER TABLE aws_settings DROP CONSTRAINT IF EXISTS fk_installs_aws_settings;
ALTER TABLE gcp_settings DROP CONSTRAINT IF EXISTS fk_installs_gcp_settings;
ALTER TABLE domains DROP CONSTRAINT IF EXISTS fk_installs_domain;

-- alter all remaining UUID columns to text + remove default value
ALTER TABLE components ALTER COLUMN id TYPE TEXT;
ALTER TABLE components ALTER COLUMN id DROP DEFAULT;
ALTER TABLE deployments ALTER COLUMN id TYPE TEXT;
ALTER TABLE deployments ALTER COLUMN id DROP DEFAULT;
ALTER TABLE installs ALTER COLUMN id TYPE TEXT;
ALTER TABLE installs ALTER COLUMN id DROP DEFAULT;
ALTER TABLE aws_settings ALTER COLUMN id TYPE TEXT;
ALTER TABLE aws_settings ALTER COLUMN id DROP DEFAULT;
ALTER TABLE gcp_settings ALTER COLUMN id TYPE TEXT;
ALTER TABLE gcp_settings ALTER COLUMN id DROP DEFAULT;
ALTER TABLE domains ALTER COLUMN id TYPE TEXT;
ALTER TABLE domains ALTER COLUMN id DROP DEFAULT;

ALTER TABLE deployments ALTER COLUMN component_id TYPE TEXT;
ALTER TABLE aws_settings ALTER COLUMN install_id TYPE TEXT;
ALTER TABLE gcp_settings ALTER COLUMN install_id TYPE TEXT;
ALTER TABLE domains ALTER COLUMN install_id TYPE TEXT;

-- re-create constraints
ALTER TABLE deployments ADD CONSTRAINT fk_components_deployments FOREIGN KEY (component_id) REFERENCES components(id);
ALTER TABLE aws_settings ADD CONSTRAINT fk_installs_aws_settings FOREIGN KEY (install_id) REFERENCES installs(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE gcp_settings ADD CONSTRAINT fk_installs_gcp_settings FOREIGN KEY (install_id) REFERENCES installs(id);
ALTER TABLE domains ADD CONSTRAINT fk_installs_domain FOREIGN KEY (install_id) REFERENCES installs(id);

-- +goose StatementEnd
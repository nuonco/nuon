-- +goose Up
-- +goose StatementBegin
CREATE TABLE github_configs (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	component_id uuid NULL,
	repo text NULL,
	directory text NULL,
	repo_owner text NULL,
	branch text NULL,
	CONSTRAINT github_configs_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_github_configs_deleted_at ON github_configs USING btree (deleted_at);

-- github_configs foreign_key
ALTER TABLE github_configs ADD	CONSTRAINT fk_components_github_config FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE github_configs;
-- +goose StatementEnd

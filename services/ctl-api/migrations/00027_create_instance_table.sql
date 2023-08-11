-- +goose Up
-- +goose StatementBegin
-- instances definition this is just a reference table between deployments,
-- components and installs

CREATE TABLE instances (
  id text not NULL,
	install_id text NULL,
	build_id text NULL,
  component_id text NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT instances_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_instances_deleted_at ON instances USING btree (deleted_at);

-- public.instances foreign keys
ALTER TABLE instances ADD CONSTRAINT fk_components_instances FOREIGN KEY (component_id) REFERENCES components(id);
ALTER TABLE instances ADD CONSTRAINT fk_installs_instances FOREIGN KEY (install_id) REFERENCES installs(id) ON DELETE CASCADE;
ALTER TABLE instances ADD CONSTRAINT fk_builds_instances FOREIGN KEY (build_id) REFERENCES builds(id);


ALTER TABLE instances ADD CONSTRAINT unq_build_deploy UNIQUE (install_id, component_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE instances;
-- +goose StatementEnd

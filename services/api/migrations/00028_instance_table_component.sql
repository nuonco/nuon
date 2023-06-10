-- +goose Up
-- +goose StatementBegin
ALTER TABLE instances DROP CONSTRAINT unq_build_deploy;

ALTER TABLE instances ADD COLUMN component_id text NULL;
ALTER TABLE instances ADD CONSTRAINT fk_components_instances FOREIGN KEY (component_id) REFERENCES components(id);

ALTER TABLE instances ADD CONSTRAINT unq_build_deploy UNIQUE (install_id, component_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE instances DROP CONSTRAINT fk_components_instances;

ALTER TABLE instances DROP COLUMN component_id;
-- +goose StatementEnd

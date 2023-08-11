-- +goose Up
-- +goose StatementBegin
-- deployments definition

CREATE TABLE deployments (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	component_id uuid NULL,
	created_by_id text NULL,
	commit_hash text NULL,
	commit_author text NULL,
	CONSTRAINT deployments_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_deployments_deleted_at ON deployments USING btree (deleted_at);

-- deployments foreign keys
ALTER TABLE deployments ADD CONSTRAINT fk_components_deployments FOREIGN KEY (component_id) REFERENCES components(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE deployments;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
-- components definition

CREATE TABLE components (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	"name" text NULL,
	app_id uuid NULL,
  created_by_id text NULL,
	build_image text NULL,
	"type" text NULL,
	CONSTRAINT components_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_components_deleted_at ON components USING btree (deleted_at);

-- forgeign key contraints
ALTER TABLE components ADD CONSTRAINT fk_apps_components FOREIGN KEY (app_id) REFERENCES apps(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE components;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
-- installs definition

-- Drop table
CREATE TABLE installs (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	created_by_id text NULL,
	"name" text NULL,
	app_id uuid NULL,
	CONSTRAINT installs_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_installs_deleted_at ON installs USING btree (deleted_at);

-- installs foreign keys
ALTER TABLE installs ADD CONSTRAINT fk_apps_installs FOREIGN KEY (app_id) REFERENCES apps(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE installs;
-- +goose StatementEnd

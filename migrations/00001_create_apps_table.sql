-- +goose Up
-- +goose StatementBegin
-- apps definition

CREATE TABLE apps (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	created_by_id text NULL,
	"name" text NULL,
	slug text NULL,
	org_id text NULL,
	github_install_id text NULL,
	CONSTRAINT apps_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_apps_deleted_at ON apps USING btree (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop table

DROP TABLE apps;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE sandbox_versions (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	sandbox_name text NULL,
	sandbox_version text NULL,
	tf_version text NULL,
	CONSTRAINT sandbox_versions_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_sandbox_versions_deleted_at ON sandbox_versions USING btree (deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sandbox_versions;
-- +goose StatementEnd

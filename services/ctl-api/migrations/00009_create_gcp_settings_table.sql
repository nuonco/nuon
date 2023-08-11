-- +goose Up
-- +goose StatementBegin
-- gcp_settings definition

CREATE TABLE gcp_settings (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	install_id uuid NULL,
	CONSTRAINT gcp_settings_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_gcp_settings_deleted_at ON gcp_settings USING btree (deleted_at);

-- gcp_settings foreign key
ALTER TABLE gcp_settings ADD CONSTRAINT fk_installs_gcp_settings FOREIGN KEY (install_id) REFERENCES installs(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE gcp_settings;
-- +goose StatementEnd

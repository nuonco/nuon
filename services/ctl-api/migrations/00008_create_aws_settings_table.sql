-- +goose Up
-- +goose StatementBegin
-- aws_settings definition

-- Drop table
CREATE TABLE aws_settings (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	install_id uuid NULL,
	region text NULL,
	iam_role_arn text NULL,
	account_id text NULL,
	CONSTRAINT aws_settings_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_aws_settings_deleted_at ON aws_settings USING btree (deleted_at);

-- aws_settings foreign keys
ALTER TABLE aws_settings ADD CONSTRAINT fk_installs_aws_settings FOREIGN KEY (install_id) REFERENCES installs(id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE aws_settings;
-- +goose StatementEnd

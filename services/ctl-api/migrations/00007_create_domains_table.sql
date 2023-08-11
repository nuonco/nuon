-- +goose Up
-- +goose StatementBegin
-- public.domains definition

-- Drop table
CREATE TABLE domains (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	install_id uuid NULL,
	"domain" text NULL,
	auto_generated bool NULL,
	nameservers text NULL,
	CONSTRAINT domains_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_domains_deleted_at ON domains USING btree (deleted_at);

-- public.domains foreign keys
ALTER TABLE domains ADD CONSTRAINT fk_installs_domain FOREIGN KEY (install_id) REFERENCES installs(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE domains;
-- +goose StatementEnd

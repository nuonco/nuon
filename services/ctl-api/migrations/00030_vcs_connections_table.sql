-- +goose Up
-- +goose StatementBegin
CREATE TABLE vcs_connections (
	id text,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	created_by_id text NULL,
	org_id text NULL,
	github_install_id text NULL,
	CONSTRAINT vcs_connections_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_vcs_connections_deleted_at ON vcs_connections USING btree (deleted_at);
ALTER TABLE vcs_connections ADD CONSTRAINT fk_vcs_connections_org FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vcs_connections DROP CONSTRAINT fk_vcs_connections_orgs;

DROP TABLE vcs_connections;
-- +goose StatementEnd

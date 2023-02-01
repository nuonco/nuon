-- +goose Up
-- +goose StatementBegin
-- user_orgs definition

CREATE TABLE user_orgs (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	user_id text NULL,
	org_id uuid NOT NULL,
	CONSTRAINT user_orgs_pkey PRIMARY KEY (id, org_id)
);
CREATE INDEX idx_user_orgs_deleted_at ON user_orgs USING btree (deleted_at);

-- public.user_orgs foreign key
ALTER TABLE user_orgs ADD CONSTRAINT fk_user_orgs_org FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_orgs;
-- +goose StatementEnd

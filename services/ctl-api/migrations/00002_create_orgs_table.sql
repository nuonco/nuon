-- +goose Up
-- +goose StatementBegin
-- public.orgs definition
-- -- public.orgs definition

-- Drop table

-- DROP TABLE public.orgs;

CREATE TABLE orgs (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	created_by_id text NULL,
	slug text NULL,
	"name" text NULL,
	CONSTRAINT orgs_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_orgs_deleted_at ON orgs USING btree (deleted_at);
CREATE UNIQUE INDEX idx_orgs_name ON orgs USING btree (name);
CREATE UNIQUE INDEX idx_orgs_slug ON orgs USING btree (slug);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orgs;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
ALTER TABLE "apps" DROP COLUMN IF EXISTS "github_install_id";
-- +goose StatementEnd

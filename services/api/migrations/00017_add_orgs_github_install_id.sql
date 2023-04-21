-- +goose Up
ALTER TABLE "orgs" ADD COLUMN "github_install_id" TEXT NULL;

-- +goose Down
ALTER TABLE "orgs" DROP COLUMN "github_install_id";
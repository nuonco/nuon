-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE components DROP COLUMN build_image;
ALTER TABLE components DROP COLUMN type;
DROP TABLE github_configs;
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

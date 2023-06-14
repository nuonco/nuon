-- +goose Up
-- +goose StatementBegin
ALTER TABLE builds ADD COLUMN instance_id text NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE builds DROP COLUMN instance_id;
-- +goose StatementEnd

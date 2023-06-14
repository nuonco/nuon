-- +goose Up
-- +goose StatementBegin
ALTER TABLE deploys ADD COLUMN instance_id text NULL;
ALTER TABLE deploys ADD CONSTRAINT fk_instances_deploys FOREIGN KEY (instance_id) REFERENCES instances(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE deploys DROP CONSTRAINT fk_instances_deploys;

ALTER TABLE deploys DROP COLUMN instance_id;
-- +goose StatementEnd

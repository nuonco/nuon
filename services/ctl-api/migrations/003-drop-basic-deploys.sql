DROP TABLE basic_deploy_configs;
ALTER TABLE docker_build_component_configs DROP COLUMN sync_only;
ALTER TABLE external_image_component_configs DROP COLUMN sync_only;

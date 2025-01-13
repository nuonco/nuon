DROP VIEW IF EXISTS runner_settings_v1;

CREATE
OR REPLACE VIEW runner_settings_v1 AS
  SELECT runner_group_settings.*, runners.id as runner_id from runner_group_settings 
	JOIN runners ON 
	runners.runner_group_id=runner_group_settings.runner_group_id 

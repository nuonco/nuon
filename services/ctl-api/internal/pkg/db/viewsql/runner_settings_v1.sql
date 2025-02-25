  SELECT runner_group_settings.*, runners.id as runner_id from runner_group_settings 
	JOIN runners ON 
	runners.runner_group_id=runner_group_settings.runner_group_id 

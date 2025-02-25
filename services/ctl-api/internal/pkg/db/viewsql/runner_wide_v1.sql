  SELECT runner_group_settings.*, runner_groups.owner_id, runner_groups.owner_type, runners.id as runner_id from runner_groups
	JOIN runner_group_settings ON 
	runner_group_settings.runner_group_id=runner_groups.id 
	JOIN runners ON 
	runners.runner_group_id=runner_group_settings.runner_group_id 

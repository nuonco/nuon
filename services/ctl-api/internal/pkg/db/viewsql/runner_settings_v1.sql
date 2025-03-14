SELECT
    runners.id AS runner_id,
    runner_group_settings.*
FROM
    runner_group_settings
    JOIN runners ON runners.runner_group_id = runner_group_settings.runner_group_id

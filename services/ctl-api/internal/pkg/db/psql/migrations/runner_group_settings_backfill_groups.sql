WITH sq AS (
    -- id and owner type
    SELECT
        runner_group_settings.id,
        runner_groups.owner_type AS owner_type
    FROM
        runner_group_settings
        JOIN runner_groups ON runner_group_settings.runner_group_id = runner_groups.id
    WHERE
        groups = '{}' :: text []
)
UPDATE
    runner_group_settings
SET
    groups = (
        -- use a case to set groups based on sq.owner_type = 'org' where sq.id  = runner_group_settings.id
        CASE
            WHEN sq.owner_type = 'orgs'     THEN ARRAY ['operations', 'sync', 'build', 'sandbox', 'runner'] :: text []
            WHEN sq.owner_type = 'installs' THEN ARRAY ['operations', 'sync', 'deploys', 'actions'] :: text []
            ELSE groups
        END
    )
FROM
    sq
WHERE
    runner_group_settings.id = sq.id;

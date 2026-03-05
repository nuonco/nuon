/* Build the app_branch_configs view with config_number */
SELECT
    abc.*,
    row_number() OVER (
        PARTITION BY app_branch_id
        ORDER BY
            abc.created_at
    ) AS config_number
FROM
    app_branch_configs abc
WHERE
    abc.deleted_at = 0

/* Load the most recent sandbox run */
WITH terraform_states_partitioned AS (
    SELECT
        ts.terraform_workspace_id,
        ROW_NUMBER() OVER (
            PARTITION BY ts.terraform_workspace_id
            ORDER BY
                ts.created_at DESC
        ) AS revision
    FROM
        terraform_states ts
    WHERE
        ts.deleted_at = 0
)

/* Build the final installs table */
SELECT
    ts.*,
    tsp.revision
FROM
    terraform_states ts FULL
    OUTER JOIN terraform_states_partitioned tsp ON tsp.terraform_workspace_id = ts.id

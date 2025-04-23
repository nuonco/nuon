/* Load the most recent sandbox run */
WITH terraform_states_partitioned AS (
    SELECT
        ts.id,
        ROW_NUMBER() OVER (
            PARTITION BY ts.terraform_workspace_id
            ORDER BY
                ts.created_at ASC
        ) AS revision
    FROM
        terraform_states ts
)

/* Build the final installs table */
SELECT
    ts.*,
    tsp.revision
FROM
    terraform_states ts
    JOIN terraform_states_partitioned tsp ON tsp.id = ts.id

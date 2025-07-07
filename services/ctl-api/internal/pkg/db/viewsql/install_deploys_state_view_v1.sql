/* install deploys with execution/deploy counts */
WITH install_deploys_with_count AS (
    SELECT
        rje.*,
        ROW_NUMBER() OVER (
            PARTITION BY rje.install_component_id
            ORDER BY
                rje.created_at DESC
        ) AS execution_number
    FROM
        install_deploys rje
    WHERE
        STATUS IN ('active', 'inactive')
)
SELECT
    rje.*
FROM
    install_deploys_with_count rje
WHERE
    rje.execution_number = 1

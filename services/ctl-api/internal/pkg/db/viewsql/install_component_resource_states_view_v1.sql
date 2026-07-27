WITH ranked_resource_states AS (
    SELECT
        s.*,
        ROW_NUMBER() OVER (
            PARTITION BY s.install_id,
            s.install_component_id,
            s.provider,
            s.api_group,
            s.kind,
            s.namespace,
            s.name
            ORDER BY
                s.observed_at DESC
        ) AS row_num
    FROM
        install_component_resource_states AS s
)
SELECT
    *
FROM
    ranked_resource_states
WHERE
    row_num = 1;

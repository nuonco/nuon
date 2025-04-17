SELECT
    row_number() OVER (
        PARTITION BY component_id
        ORDER BY
            created_at
    ) AS version,
    *
FROM
    component_config_connections;

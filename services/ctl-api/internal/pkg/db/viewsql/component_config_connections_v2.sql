SELECT
    ccc.*,
    (
        SELECT n FROM (
            SELECT ccc2.id AS id,
                   row_number() OVER (ORDER BY ccc2.created_at, ccc2.id) AS n
            FROM component_config_connections ccc2
            WHERE ccc2.component_id = ccc.component_id
        ) sub
        WHERE sub.id = ccc.id
    ) AS version,
    (
        SELECT n FROM (
            SELECT ac2.id AS id,
                   row_number() OVER (ORDER BY ac2.created_at, ac2.id) AS n
            FROM app_configs ac2
            WHERE ac2.app_id = ac.app_id
        ) sub
        WHERE sub.id = ac.id
    ) AS app_config_version
FROM component_config_connections ccc
JOIN app_configs ac ON ac.id = ccc.app_config_id

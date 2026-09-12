SELECT DISTINCT ON (ccc.component_id)
    ccc.*,
    (
        SELECT n FROM (
            SELECT ccc2.id AS id,
                   row_number() OVER (ORDER BY ccc2.created_at, ccc2.id) AS n
            FROM component_config_connections ccc2
            WHERE ccc2.component_id = ccc.component_id
        ) sub
        WHERE sub.id = ccc.id
    ) AS version
FROM component_config_connections ccc
ORDER BY ccc.component_id, ccc.created_at DESC, ccc.id DESC

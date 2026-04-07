SELECT ccc.*
FROM component_config_connections ccc
INNER JOIN (
    SELECT component_id, MAX(version) AS max_version
    FROM component_config_connections
    WHERE deleted_at = 0
    GROUP BY component_id
) latest ON ccc.component_id = latest.component_id AND ccc.version = latest.max_version
WHERE ccc.deleted_at = 0

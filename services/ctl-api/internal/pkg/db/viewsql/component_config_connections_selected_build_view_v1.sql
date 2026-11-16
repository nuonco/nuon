SELECT
    c.id AS component_config_connection_id,
    c.app_config_id,
    c.component_id,
    c.checksum,
    COALESCE(pin.id, legacy.id) AS selected_build_id
FROM
    component_config_connections c
    JOIN app_configs a ON a.id = c.app_config_id
    LEFT JOIN component_builds pin ON pin.id = c.latest_build_id
        AND pin.deleted_at = 0
        AND COALESCE(NULLIF(pin.status_v2 ->> 'status', ''), pin.status) = 'active'
        AND EXISTS (
            SELECT 1
            FROM component_config_connections bc
            WHERE bc.id = pin.component_config_connection_id
                AND bc.deleted_at = 0
                AND bc.component_id = c.component_id
        )
    LEFT JOIN LATERAL (
        SELECT b.id
        FROM component_config_connections old
            JOIN component_builds b ON b.component_config_connection_id = old.id
        WHERE pin.id IS NULL
            AND NOT a.builds_pinned
            AND b.deleted_at = 0
            AND b.status = 'active'
            AND old.component_id = c.component_id
            AND CASE WHEN COALESCE(c.checksum, '') = '' THEN old.id = c.id ELSE old.checksum = c.checksum END
        ORDER BY b.created_at DESC, b.id
        LIMIT 1
    ) legacy ON true
WHERE
    c.deleted_at = 0

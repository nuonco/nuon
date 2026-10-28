(
    SELECT COALESCE((ARRAY['unhealthy', 'degraded', 'other', 'unknown', 'progressing', 'healthy'])[MIN(
        CASE
            WHEN component_health.health IS NULL OR component_health.health IN ('', 'unset', 'not-applicable') THEN NULL
            WHEN component_health.health = 'unhealthy' THEN 1
            WHEN component_health.health = 'degraded' THEN 2
            WHEN component_health.health = 'unknown' THEN 4
            WHEN component_health.health = 'progressing' THEN 5
            WHEN component_health.health = 'healthy' THEN 6
            ELSE 3
        END)], 'unset')
    FROM (
        SELECT COALESCE(NULLIF(ic.health_status_v2->>'status', ''), ic.health_status) AS health
        FROM install_components ic
        WHERE ic.deleted_at = 0 AND ic.install_id = {{install_id}}
    ) component_health
)

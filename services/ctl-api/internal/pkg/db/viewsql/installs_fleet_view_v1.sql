SELECT
    i.id,
    i.org_id,
    i.app_id,
    i.name,
    i.labels,
    i.app_config_id,
    COALESCE(i.app_config_ref->>'applied_config_id', '') AS applied_config_id,
    COALESCE(conn.app_branch_id, '') AS app_branch_id,
    COALESCE(conn.app_branch_group, '') AS app_branch_group,
    COALESCE(branch.name, '') AS app_branch_name,
    health.status AS composite_health_status,
    COALESCE(runner.status, 'deprovisioned') AS runner_status
FROM installs i
LEFT JOIN LATERAL (
    SELECT c.app_branch_id, c.app_branch_group
    FROM install_app_branch_connections c
    WHERE c.install_id = i.id
      AND c.active = true
      AND c.deleted_at = 0
    ORDER BY c.created_at DESC, c.id DESC
    LIMIT 1
) conn ON true
LEFT JOIN app_branches branch
    ON branch.id = conn.app_branch_id
   AND branch.deleted_at = 0
LEFT JOIN LATERAL (
    SELECT
        CASE
            WHEN count(*) FILTER (WHERE ic.health_status = 'unhealthy') > 0 THEN 'unhealthy'
            WHEN count(*) FILTER (WHERE ic.health_status = 'degraded') > 0 THEN 'degraded'
            WHEN count(*) FILTER (WHERE ic.health_status = 'unknown') > 0 THEN 'unknown'
            WHEN count(*) FILTER (WHERE ic.health_status = 'progressing') > 0 THEN 'progressing'
            WHEN count(*) FILTER (WHERE ic.health_status = 'healthy') > 0 THEN 'healthy'
            ELSE ''
        END AS status
    FROM install_components ic
    WHERE ic.install_id = i.id
      AND ic.deleted_at = 0
) health ON true
LEFT JOIN LATERAL (
    SELECT r.status
    FROM runner_groups g
    JOIN runners r
      ON r.runner_group_id = g.id
     AND r.deleted_at = 0
    WHERE g.owner_id = i.id
      AND g.owner_type = 'installs'
      AND g.deleted_at = 0
    ORDER BY r.created_at DESC
    LIMIT 1
) runner ON true
WHERE i.deleted_at = 0

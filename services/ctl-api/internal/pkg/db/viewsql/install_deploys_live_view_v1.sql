/* the deploy each component is actually running: its applied pointer, or for legacy
   components without one, the newest successful real (non plan-only) deploy, excluding
   recover rows and any teardown that removed the component */
SELECT
    live.*
FROM (
    SELECT
        d.*,
        ic.install_id
    FROM
        install_components ic
        JOIN install_deploys d ON d.id = ic.app_config_ref ->> 'applied_config_by_id'
            AND d.install_component_id = ic.id
    WHERE
        d.deleted_at = 0
        AND ic.app_config_ref ->> 'applied_config_by_type' = 'install_deploys'
        AND COALESCE(ic.app_config_ref ->> 'applied_config_id', '') <> ''
    UNION ALL
    (SELECT DISTINCT ON (ic.install_id, d.install_component_id)
        d.*,
        ic.install_id
    FROM
        install_components ic
        JOIN install_deploys d ON d.install_component_id = ic.id
    WHERE
        COALESCE(ic.app_config_ref ->> 'applied_config_by_id', '') = ''
        AND d.deleted_at = 0
        AND d.type IS DISTINCT FROM 'recover'
        AND (
            d.applied_at IS NOT NULL
            OR (
                d.type = 'sync-image'
                AND COALESCE(NULLIF(d.status_v2 ->> 'status', ''), d.status) = 'active'
            )
        )
        AND NOT EXISTS (
            SELECT 1
            FROM install_workflows w
            WHERE w.id = d.install_workflow_id AND w.plan_only = true
        )
    ORDER BY
        ic.install_id,
        d.install_component_id,
        COALESCE(d.applied_at, d.updated_at) DESC,
        d.created_at DESC)
) live
WHERE
    live.type IS DISTINCT FROM 'teardown'

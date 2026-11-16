SELECT
    drifted.*
FROM (
    SELECT
        'install_component' AS resource_type,
        ic.org_id,
        ic.install_id,
        ic.id AS install_component_id,
        NULL AS install_sandbox_id,
        ic.component_id,
        c.name AS component_name,
        sel.component_config_connection_id AS expected_config_id,
        sel.selected_build_id AS expected_build_id,
        live.id AS live_id,
        live.component_build_id AS live_build_id,
        live.applied_at AS live_applied_at,
        live.install_workflow_id AS live_workflow_id,
        live.id IS NULL AND sel.selected_build_id IS NOT NULL AS not_deployed,
        live.id IS NOT NULL AND sel.selected_build_id IS DISTINCT FROM live.component_build_id AS build_changed,
        live.id IS NOT NULL AND sel.checksum IS DISTINCT FROM live_config.checksum AS config_changed,
        scan.id AS drift_scan_id,
        COALESCE(scan.status = 'drifted', false)
            AND sel.selected_build_id IS NOT DISTINCT FROM live.component_build_id
            AND sel.checksum IS NOT DISTINCT FROM live_config.checksum AS state_drift,
        CASE WHEN last_wf.finished_at IS NULL
            AND COALESCE(last_wf.status ->> 'status', '') IN ('', 'pending', 'queued', 'in-progress', 'retrying', 'approval-awaiting', 'approved', 'failed-pending-retry')
            THEN last.id END AS open_deploy_id,
        CASE WHEN last_wf.finished_at IS NULL
            AND COALESCE(last_wf.status ->> 'status', '') IN ('', 'pending', 'queued', 'in-progress', 'retrying', 'approval-awaiting', 'approved', 'failed-pending-retry')
            THEN last_wf.id END AS open_workflow_id
    FROM
        install_components ic
        JOIN installs i ON i.id = ic.install_id
        JOIN components c ON c.id = ic.component_id
        LEFT JOIN component_config_connections expected_config ON expected_config.app_config_id = i.app_config_id
            AND expected_config.component_id = ic.component_id
            AND expected_config.deleted_at = 0
        LEFT JOIN component_config_connections_selected_build_view_v1 sel ON sel.component_config_connection_id = expected_config.id
        LEFT JOIN install_deploys_live_view_v1 live ON live.install_id = ic.install_id
            AND live.install_component_id = ic.id
        LEFT JOIN component_builds live_build ON live_build.id = live.component_build_id
        LEFT JOIN component_config_connections live_config ON live_config.id = live_build.component_config_connection_id
        LEFT JOIN LATERAL (
            SELECT d.id, d.created_at, d.install_workflow_id
            FROM install_deploys d
                LEFT JOIN install_workflows w ON w.id = d.install_workflow_id
            WHERE d.install_component_id = ic.id
                AND d.deleted_at = 0
                AND w.plan_only IS NOT TRUE
            ORDER BY d.created_at DESC
            LIMIT 1
        ) last ON true
        LEFT JOIN install_workflows last_wf ON last_wf.id = last.install_workflow_id
        LEFT JOIN LATERAL (
            SELECT d.id, d.status
            FROM install_deploys d
                JOIN install_workflows w ON w.id = d.install_workflow_id AND w.plan_only = true
            WHERE live.id IS NOT NULL
                AND d.install_component_id = ic.id
                AND d.deleted_at = 0
                AND d.component_build_id = live.component_build_id
                AND d.created_at >= COALESCE(live.applied_at, live.created_at)
                AND d.created_at >= COALESCE(last.created_at, '-infinity')
            ORDER BY d.created_at DESC
            LIMIT 1
        ) scan ON true
    WHERE
        ic.deleted_at = 0
    UNION ALL
    SELECT
        'install_sandbox' AS resource_type,
        i.org_id,
        isr.install_id,
        NULL AS install_component_id,
        isr.install_sandbox_id,
        NULL AS component_id,
        NULL AS component_name,
        NULL AS expected_config_id,
        NULL AS expected_build_id,
        NULL AS live_id,
        NULL AS live_build_id,
        NULL::timestamptz AS live_applied_at,
        NULL AS live_workflow_id,
        false AS not_deployed,
        false AS build_changed,
        false AS config_changed,
        isr.id AS drift_scan_id,
        true AS state_drift,
        NULL AS open_deploy_id,
        NULL AS open_workflow_id
    FROM
        install_sandbox_runs isr
        JOIN installs i ON i.id = isr.install_id
    WHERE
        isr.status = 'drifted'
        AND NOT EXISTS (
            SELECT 1
            FROM install_sandbox_runs sr2
            WHERE sr2.install_sandbox_id = isr.install_sandbox_id
                AND (sr2.created_at > isr.created_at OR (sr2.created_at = isr.created_at AND sr2.id > isr.id))
        )
) drifted
WHERE
    drifted.not_deployed
    OR drifted.build_changed
    OR drifted.config_changed
    OR drifted.state_drift

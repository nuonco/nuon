SELECT
    drifted.*,
    drifted.not_deployed OR drifted.build_changed OR drifted.config_changed OR drifted.state_drift AS out_of_sync
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
        CASE WHEN scan.at_live_build THEN scan.id END AS drift_scan_id,
        COALESCE(scan.at_live_build AND scan.status = 'drifted', false)
            AND sel.selected_build_id IS NOT DISTINCT FROM live.component_build_id
            AND sel.checksum IS NOT DISTINCT FROM live_config.checksum AS state_drift,
        CASE WHEN last_wf.finished_at IS NULL
            AND COALESCE(last_wf.status ->> 'status', '') IN ('', 'pending', 'queued', 'in-progress', 'retrying', 'approval-awaiting', 'approved', 'failed-pending-retry')
            THEN last.id END AS open_deploy_id,
        CASE WHEN last_wf.finished_at IS NULL
            AND COALESCE(last_wf.status ->> 'status', '') IN ('', 'pending', 'queued', 'in-progress', 'retrying', 'approval-awaiting', 'approved', 'failed-pending-retry')
            THEN last_wf.id END AS open_workflow_id,
        last.component_build_id AS last_build_id,
        last.type AS last_type,
        COALESCE(NULLIF(last.status_v2 ->> 'status', ''), last.status) AS last_status,
        last.planned_at AS last_planned_at,
        last.planned_inputs_digest AS last_planned_inputs_digest,
        last.status_v2 -> 'metadata' ->> 'policy_action' AS last_policy_action,
        last_wf.id AS last_workflow_id,
        last_wf.type AS last_workflow_type,
        last_wf.status ->> 'status' AS last_workflow_status,
        last_wf.created_at AS last_workflow_created_at,
        COALESCE(scan.at_live_build AND scan.status = 'drifted', false) AS scan_drifted,
        live_wf.type AS live_workflow_type,
        live_wf.metadata -> 'app_branch_run_id' AS live_app_branch_run_id,
        real_plan.id AS real_plan_approval_id,
        scan_plan.id AS scan_plan_approval_id,
        NULL::timestamptz AS last_created_at,
        NULL AS last_app_sandbox_config_id,
        NULL AS scan_id,
        NULL AS scan_status,
        NULL AS scan_app_sandbox_config_id,
        NULL::timestamptz AS scan_created_at
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
            AND live.org_id = ic.org_id
        LEFT JOIN install_workflows live_wf ON live_wf.id = live.install_workflow_id AND live_wf.deleted_at = 0
        LEFT JOIN component_builds live_build ON live_build.id = live.component_build_id
        LEFT JOIN component_config_connections live_config ON live_config.id = live_build.component_config_connection_id
        LEFT JOIN LATERAL (
            SELECT d.id, d.created_at, d.install_workflow_id, d.component_build_id, d.type, d.status, d.status_v2, d.planned_at, d.planned_inputs_digest
            FROM install_deploys d
                LEFT JOIN install_workflows w ON w.id = d.install_workflow_id
            WHERE d.install_component_id = ic.id
                AND d.deleted_at = 0
                AND d.plan_only = false
                AND w.plan_only IS NOT TRUE
            ORDER BY d.created_at DESC
            LIMIT 1
        ) last ON true
        LEFT JOIN install_workflows last_wf ON last_wf.id = last.install_workflow_id AND last_wf.deleted_at = 0
        LEFT JOIN LATERAL (
            SELECT d.id, d.status, d.component_build_id = live.component_build_id AS at_live_build
            FROM install_deploys d
                JOIN install_workflows w ON w.id = d.install_workflow_id AND w.plan_only = true
            WHERE live.id IS NOT NULL
                AND d.install_component_id = ic.id
                AND d.deleted_at = 0
                AND d.created_at >= COALESCE(live.applied_at, live.created_at)
                AND d.created_at >= COALESCE(last.created_at, '-infinity')
            ORDER BY d.created_at DESC
            LIMIT 1
        ) scan ON true
        LEFT JOIN LATERAL (
            SELECT a.id
            FROM install_workflow_step_approvals a
            WHERE a.owner_id = last.id
                AND a.owner_type = 'install_deploys'
                AND a.deleted_at = 0
            ORDER BY a.created_at DESC
            LIMIT 1
        ) real_plan ON true
        LEFT JOIN LATERAL (
            SELECT a.id
            FROM install_workflow_step_approvals a
            WHERE scan.at_live_build
                AND a.owner_id = scan.id
                AND a.owner_type = 'install_deploys'
                AND a.deleted_at = 0
            ORDER BY a.created_at DESC
            LIMIT 1
        ) scan_plan ON true
    WHERE
        ic.deleted_at = 0
    UNION ALL
    SELECT
        'install_sandbox' AS resource_type,
        i.org_id,
        sb.install_id,
        NULL AS install_component_id,
        sb.id AS install_sandbox_id,
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
        CASE WHEN newest.status = 'drifted' THEN newest.id END AS drift_scan_id,
        newest.status = 'drifted' AS state_drift,
        NULL AS open_deploy_id,
        NULL AS open_workflow_id,
        NULL AS last_build_id,
        NULL AS last_type,
        COALESCE(NULLIF(last.status_v2 ->> 'status', ''), last.status) AS last_status,
        last.planned_at AS last_planned_at,
        last.planned_inputs_digest AS last_planned_inputs_digest,
        last.status_v2 -> 'metadata' ->> 'policy_action' AS last_policy_action,
        NULL AS last_workflow_id,
        NULL AS last_workflow_type,
        NULL AS last_workflow_status,
        NULL::timestamptz AS last_workflow_created_at,
        newest.status = 'drifted' AS scan_drifted,
        live_wf.type AS live_workflow_type,
        live_wf.metadata -> 'app_branch_run_id' AS live_app_branch_run_id,
        real_plan.id AS real_plan_approval_id,
        scan_plan.id AS scan_plan_approval_id,
        last.created_at AS last_created_at,
        last.app_sandbox_config_id AS last_app_sandbox_config_id,
        scan.id AS scan_id,
        COALESCE(NULLIF(scan.status_v2 ->> 'status', ''), scan.status) AS scan_status,
        scan.app_sandbox_config_id AS scan_app_sandbox_config_id,
        scan.created_at AS scan_created_at
    FROM
        install_sandboxes sb
        JOIN installs i ON i.id = sb.install_id
        LEFT JOIN install_sandbox_runs_live_view_v1 live ON live.install_sandbox_id = sb.id
            AND live.org_id = i.org_id
        LEFT JOIN install_workflows live_wf ON live_wf.id = live.install_workflow_id AND live_wf.deleted_at = 0
        LEFT JOIN LATERAL (
            SELECT r.id, r.created_at, r.status, r.status_v2, r.app_sandbox_config_id, r.planned_at, r.planned_inputs_digest
            FROM install_sandbox_runs r
                LEFT JOIN install_workflows w ON w.id = r.install_workflow_id AND w.deleted_at = 0
            WHERE r.install_sandbox_id = sb.id
                AND r.deleted_at = 0
                AND w.plan_only IS NOT TRUE
            ORDER BY r.created_at DESC
            LIMIT 1
        ) last ON true
        LEFT JOIN LATERAL (
            SELECT r.id, r.created_at, r.status, r.status_v2, r.app_sandbox_config_id
            FROM install_sandbox_runs r
                JOIN install_workflows w ON w.id = r.install_workflow_id AND w.deleted_at = 0 AND w.plan_only = true
            WHERE r.install_sandbox_id = sb.id
                AND r.deleted_at = 0
            ORDER BY r.created_at DESC
            LIMIT 1
        ) scan ON true
        LEFT JOIN LATERAL (
            SELECT a.id
            FROM install_workflow_step_approvals a
            WHERE last.id IS NOT NULL
                AND a.owner_id = last.id
                AND a.owner_type = 'install_sandbox_runs'
                AND a.deleted_at = 0
            ORDER BY a.created_at DESC
            LIMIT 1
        ) real_plan ON true
        LEFT JOIN LATERAL (
            SELECT a.id
            FROM install_workflow_step_approvals a
            WHERE scan.id IS NOT NULL
                AND a.owner_id = scan.id
                AND a.owner_type = 'install_sandbox_runs'
                AND a.deleted_at = 0
            ORDER BY a.created_at DESC
            LIMIT 1
        ) scan_plan ON true
        LEFT JOIN LATERAL (
            SELECT r.id, r.status
            FROM install_sandbox_runs r
            WHERE r.install_sandbox_id = sb.id
            ORDER BY r.created_at DESC, r.id DESC
            LIMIT 1
        ) newest ON true
) drifted

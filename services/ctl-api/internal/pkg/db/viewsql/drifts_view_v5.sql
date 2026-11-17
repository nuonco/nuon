SELECT
    drifted.*,
    drifted.expected_state = 'ready' AND (drifted.live_state = 'not_deployed' OR drifted.live_outdated) AS out_of_sync
FROM (
    SELECT
        'install_component' AS resource_type,
        ic.org_id,
        ic.install_id,
        ic.id AS install_component_id,
        NULL AS install_sandbox_id,
        ic.component_id,
        c.name AS component_name,
        expected_config.id AS expected_config_id,
        sel.selected_build_id AS expected_build_id,
        live.id AS live_id,
        live.component_build_id AS live_build_id,
        live.applied_at AS live_applied_at,
        live.install_workflow_id AS live_workflow_id,
        last.id AS last_deploy_id,
        CASE WHEN scan.at_live_build THEN scan.id END AS drift_check_id,
        st.expected_state,
        st.live_state,
        cmp.live_outdated,
        st.expected_state = 'ready' AND st.live_state = 'deployed' AND NOT cmp.live_outdated
            AND COALESCE(scan.at_live_build AND scan.status = 'drifted', false) AS state_drift
    FROM
        install_components ic
        JOIN installs i ON i.id = ic.install_id
        JOIN components c ON c.id = ic.component_id
        LEFT JOIN component_config_connections expected_config ON expected_config.app_config_id = i.app_config_id
            AND expected_config.component_id = ic.component_id
            AND expected_config.deleted_at = 0
        LEFT JOIN component_config_connections_selected_build_view_v1 sel ON sel.component_config_connection_id = expected_config.id
        LEFT JOIN component_builds expected_build ON expected_build.id = sel.selected_build_id
        LEFT JOIN install_deploys_live_view_v1 live ON live.install_id = ic.install_id
            AND live.install_component_id = ic.id
            AND live.org_id = ic.org_id
        LEFT JOIN component_builds live_build ON live_build.id = live.component_build_id
        LEFT JOIN component_config_connections live_config ON live_config.id = live_build.component_config_connection_id
        LEFT JOIN LATERAL (
            SELECT
                d.id,
                d.created_at,
                d.type,
                COALESCE(NULLIF(d.status_v2 ->> 'status', ''), d.status) AS status,
                w.type AS workflow_type
            FROM install_deploys d
                LEFT JOIN install_workflows w ON w.id = d.install_workflow_id
            WHERE d.install_component_id = ic.id
                AND d.deleted_at = 0
                AND d.plan_only = false
                AND w.plan_only IS NOT TRUE
            ORDER BY d.created_at DESC
            LIMIT 1
        ) last ON true
        LEFT JOIN LATERAL (
            SELECT d.id, COALESCE(NULLIF(d.status_v2 ->> 'status', ''), d.status) AS status,
                d.component_build_id = live.component_build_id AS at_live_build
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
        CROSS JOIN LATERAL (
            SELECT
                CASE
                    WHEN expected_config.id IS NULL THEN 'not_in_config'
                    WHEN sel.selected_build_id IS NULL THEN 'no_build'
                    ELSE 'ready'
                END AS expected_state,
                CASE
                    WHEN last.type = 'teardown' AND last.status = 'inactive'
                        AND COALESCE(last.workflow_type, '') NOT IN ('input_update', 'component_disabled') THEN 'torn_down'
                    WHEN live.id IS NOT NULL THEN 'deployed'
                    WHEN COALESCE(NULLIF(ic.status_v2 ->> 'status', ''), ic.status) IN ('active', 'noop', 'error')
                        OR (COALESCE(NULLIF(ic.status_v2 ->> 'status', ''), ic.status, '') NOT IN ('inactive', 'deleted', 'disabled')
                            AND COALESCE(NULLIF(ic.health_status_v2 ->> 'status', ''), ic.health_status, '') NOT IN ('', 'not-applicable')) THEN 'unknown'
                    ELSE 'not_deployed'
                END AS live_state
        ) st
        CROSS JOIN LATERAL (
            SELECT st.expected_state = 'ready' AND st.live_state = 'deployed' AND (
                COALESCE(sel.checksum, '') = ''
                OR sel.checksum IS DISTINCT FROM live_config.checksum
                OR NOT COALESCE(
                    sel.selected_build_id = live.component_build_id
                    OR (COALESCE(expected_build.source_checksum, '') <> ''
                        AND expected_build.source_checksum = live_build.source_checksum)
                    OR (expected_build.source_identity IS NULL
                        AND live_build.source_identity IS NULL
                        AND expected_build.source_digest LIKE 'sha256:%'
                        AND expected_build.source_digest = live_build.source_digest),
                    false)) AS live_outdated
        ) cmp
    WHERE
        ic.deleted_at = 0
        AND i.deleted_at = 0
    UNION ALL
    SELECT
        'install_sandbox' AS resource_type,
        i.org_id,
        sb.install_id,
        NULL AS install_component_id,
        sb.id AS install_sandbox_id,
        NULL AS component_id,
        NULL AS component_name,
        cfg.id AS expected_config_id,
        NULL AS expected_build_id,
        live.id AS live_id,
        NULL AS live_build_id,
        live.applied_at AS live_applied_at,
        live.install_workflow_id AS live_workflow_id,
        NULL AS last_deploy_id,
        CASE WHEN newest.status = 'drifted' THEN newest.id END AS drift_check_id,
        CASE WHEN cfg.id IS NULL THEN 'not_in_config' ELSE 'ready' END AS expected_state,
        CASE
            WHEN live.id IS NOT NULL THEN 'deployed'
            WHEN COALESCE(sb.app_config_ref ->> 'applied_config_by_id', '') = ''
                AND COALESCE(NULLIF(sb.status_v2 ->> 'status', ''), sb.status) = 'active' THEN 'unknown'
            ELSE 'not_deployed'
        END AS live_state,
        false AS live_outdated,
        cfg.id IS NOT NULL AND live.id IS NOT NULL AND COALESCE(newest.status = 'drifted', false) AS state_drift
    FROM
        install_sandboxes sb
        JOIN installs i ON i.id = sb.install_id AND i.deleted_at = 0
        LEFT JOIN install_sandbox_runs_live_view_v1 live ON live.install_id = sb.install_id
            AND live.install_sandbox_id = sb.id
            AND live.org_id = i.org_id
        LEFT JOIN LATERAL (
            SELECT a.id
            FROM app_sandbox_configs a
            WHERE a.app_config_id = i.app_config_id
                AND a.deleted_at = 0
            ORDER BY a.created_at DESC
            LIMIT 1
        ) cfg ON true
        LEFT JOIN LATERAL (
            SELECT r.id, COALESCE(NULLIF(r.status_v2 ->> 'status', ''), r.status) AS status
            FROM install_sandbox_runs r
            WHERE r.install_sandbox_id = sb.id
                AND r.deleted_at = 0
            ORDER BY r.created_at DESC, r.id DESC
            LIMIT 1
        ) newest ON true
    WHERE sb.deleted_at = 0
) drifted

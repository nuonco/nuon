/* per install, one sync row per deployable resource: component, sandbox, or stack.
   Columns follow the shared lifecycle rather than the resource kind, and
   resource_type says which table each id points at:
     resource_id            install_components / install_sandboxes / install_stacks
     expected_config_id     component_config_connections / app_sandbox_configs / app_configs
     expected_artifact_id   component_builds / app_sandbox_builds / install_stack_versions
     live_id                install_deploys / install_sandbox_runs / install_stack_versions
     live_artifact_id       component_builds / (unset) / (unset)
     live_config_id         (unset) / (unset) / app_configs
     attempt_id             install_deploys / (unset) / (unset)
     attempt_workflow_id    (unset) / (unset) / install_workflows
     drift_check_id         install_deploys / install_sandbox_runs / (never scanned)
   live_source is recorded when the live record was written as applied, inferred
   when it was reconstructed (legacy stacks without an applied pointer).
   Classification (in-sync / changes-not-applied / torn-down / disabled) stays in Go;
   out_of_sync is ready-but-different, state_drift is scan-detected live drift */
SELECT
    drifted.*,
    drifted.expected_state = 'ready' AND (drifted.live_state = 'not_deployed' OR drifted.live_outdated) AS out_of_sync
FROM (
    SELECT
        'install_component' AS resource_type,
        ic.org_id,
        ic.install_id,
        ic.id AS resource_id,
        c.name AS resource_name,
        expected_config.id AS expected_config_id,
        sel.selected_build_id AS expected_artifact_id,
        NULL AS expected_artifact_status,
        live.id AS live_id,
        live.component_build_id AS live_artifact_id,
        NULL AS live_config_id,
        CASE WHEN live.id IS NOT NULL THEN 'recorded' END AS live_source,
        live.applied_at AS live_applied_at,
        live.install_workflow_id AS live_workflow_id,
        last.id AS attempt_id,
        NULL AS attempt_workflow_id,
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
        LEFT JOIN LATERAL (
            SELECT ec.id
            FROM component_config_connections ec
            WHERE i.app_config_id <> ''
                AND ec.app_config_id = i.app_config_id
                AND ec.component_id = ic.component_id
                AND ec.deleted_at = 0
            ORDER BY ec.created_at DESC, ec.id
            LIMIT 1
        ) expected_config ON true
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
            ORDER BY d.created_at DESC, d.id DESC
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
                AND d.component_build_id = live.component_build_id
                AND d.created_at >= COALESCE(live.applied_at, live.created_at)
                AND d.created_at >= COALESCE(last.created_at, '-infinity')
            ORDER BY d.created_at DESC, d.id DESC
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
            SELECT
                CASE WHEN expected_build.source_identity ->> 'version' = '1'
                    AND expected_build.source_identity ->> 'coverage' = 'complete'
                    AND expected_build.source_identity ->> 'captured' = 'true'
                    AND COALESCE(expected_build.source_identity ->> 'config_checksum', '') <> ''
                    AND COALESCE(expected_build.source_identity ->> 'checksum', '') <> ''
                    AND NOT (COALESCE(expected_build.source_identity ->> 'digest', '') LIKE 'git:%'
                        AND COALESCE(expected_build.source_identity ->> 'inputs_recorded', '') <> 'true')
                THEN expected_build.source_identity ->> 'checksum' END AS expected_source,
                CASE WHEN live_build.source_identity ->> 'version' = '1'
                    AND live_build.source_identity ->> 'coverage' = 'complete'
                    AND live_build.source_identity ->> 'captured' = 'true'
                    AND COALESCE(live_build.source_identity ->> 'config_checksum', '') <> ''
                    AND COALESCE(live_build.source_identity ->> 'checksum', '') <> ''
                    AND NOT (COALESCE(live_build.source_identity ->> 'digest', '') LIKE 'git:%'
                        AND COALESCE(live_build.source_identity ->> 'inputs_recorded', '') <> 'true')
                THEN live_build.source_identity ->> 'checksum' END AS live_source
        ) eligible
        CROSS JOIN LATERAL (
            SELECT st.expected_state = 'ready' AND st.live_state = 'deployed' AND NOT COALESCE(
                sel.selected_build_id = live.component_build_id
                OR (COALESCE(expected_build.source_checksum, '') <> ''
                    AND expected_build.source_checksum = live_build.source_checksum
                    AND COALESCE(sel.checksum, '') <> ''
                    AND sel.checksum = live_config.checksum
                    AND COALESCE(eligible.expected_source, '') <> ''
                    AND eligible.expected_source = eligible.live_source
                    AND eligible.expected_source = expected_build.source_checksum)
                OR (expected_build.source_identity IS NULL
                    AND live_build.source_identity IS NULL
                    AND expected_build.source_digest LIKE 'sha256:%'
                    AND expected_build.source_digest = live_build.source_digest
                    AND COALESCE(sel.checksum, '') <> ''
                    AND sel.checksum = live_config.checksum),
                false) AS live_outdated
        ) cmp
    WHERE
        ic.deleted_at = 0
        AND i.deleted_at = 0
    UNION ALL
    SELECT
        'install_sandbox' AS resource_type,
        i.org_id,
        sb.install_id,
        sb.id AS resource_id,
        'sandbox' AS resource_name,
        cfg.id AS expected_config_id,
        CASE WHEN a.builds_pinned THEN pin.id ELSE direct.id END AS expected_artifact_id,
        NULL AS expected_artifact_status,
        live.id AS live_id,
        NULL AS live_artifact_id,
        NULL AS live_config_id,
        CASE WHEN live.id IS NOT NULL THEN 'recorded' END AS live_source,
        live.applied_at AS live_applied_at,
        live.install_workflow_id AS live_workflow_id,
        NULL AS attempt_id,
        NULL AS attempt_workflow_id,
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
        LEFT JOIN app_configs a ON a.id = i.app_config_id
        LEFT JOIN app_sandbox_builds pin ON a.builds_pinned
            AND pin.id = NULLIF(a.sandbox_build_id, '')
            AND pin.app_id = a.app_id
            AND pin.deleted_at = 0
            AND pin.status = 'active'
        LEFT JOIN LATERAL (
            SELECT b.id
            FROM app_sandbox_builds b
            WHERE NOT a.builds_pinned
                AND b.app_config_id = a.id
                AND b.status = 'active'
                AND b.deleted_at = 0
            ORDER BY b.created_at DESC
            LIMIT 1
        ) direct ON true
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
    UNION ALL
    SELECT
        'install_stack' AS resource_type,
        i.org_id,
        s.install_id,
        s.id AS resource_id,
        'stack' AS resource_name,
        desired.app_config_id AS expected_config_id,
        desired.id AS expected_artifact_id,
        desired.status AS expected_artifact_status,
        live.id AS live_id,
        NULL AS live_artifact_id,
        live.app_config_id AS live_config_id,
        CASE WHEN live.id IS NULL THEN NULL
            WHEN live.id = s.app_config_ref ->> 'applied_config_by_id' THEN 'recorded'
            ELSE 'inferred' END AS live_source,
        CASE WHEN live.id = s.app_config_ref ->> 'applied_config_by_id'
            THEN (s.app_config_ref ->> 'applied_config_at')::timestamptz END AS live_applied_at,
        NULL AS live_workflow_id,
        NULL AS attempt_id,
        desired_wf.id AS attempt_workflow_id,
        NULL AS drift_check_id,
        CASE WHEN COALESCE(i.app_config_id, '') = '' THEN 'not_in_config'
            WHEN desired.id IS NULL THEN 'no_build'
            ELSE 'ready' END AS expected_state,
        CASE
            WHEN live.id IS NOT NULL THEN 'deployed'
            WHEN desired.id IS NULL AND EXISTS (
                SELECT 1 FROM install_stack_versions x
                WHERE x.install_stack_id = s.id AND x.deleted_at = 0
            ) THEN 'unknown'
            ELSE 'not_deployed'
        END AS live_state,
        live.id IS NOT NULL AND desired.id IS NOT NULL AND live.id <> desired.id AS live_outdated,
        false AS state_drift
    FROM
        install_stacks s
        JOIN installs i ON i.id = s.install_id AND i.deleted_at = 0
        LEFT JOIN install_stack_versions_live_view_v1 live ON live.install_id = s.install_id
            AND live.install_stack_id = s.id
        LEFT JOIN LATERAL (
            SELECT COALESCE(dup.id, v.id) AS id, COALESCE(dup.app_config_id, v.app_config_id) AS app_config_id,
                COALESCE(dup.status, v.status) ->> 'status' AS status
            FROM install_stack_versions v
                LEFT JOIN install_stack_versions dup ON v.status ->> 'status' = 'outdated'
                    AND dup.id = v.status -> 'metadata' ->> 'duplicate_of_version_id'
                    AND dup.install_stack_id = s.id AND dup.deleted_at = 0
            WHERE i.app_config_id <> '' AND v.install_stack_id = s.id AND v.app_config_id = i.app_config_id AND v.deleted_at = 0
            ORDER BY v.created_at DESC, v.id
            LIMIT 1
        ) desired ON true
        LEFT JOIN LATERAL (
            SELECT w.id FROM install_workflows w
                JOIN install_workflow_steps st ON st.install_workflow_id = w.id AND st.deleted_at = 0
            WHERE st.step_target_type = 'install_stack_versions' AND st.step_target_id = desired.id
                AND w.plan_only IS NOT TRUE AND w.deleted_at = 0
            ORDER BY w.created_at DESC, w.id LIMIT 1
        ) desired_wf ON true
    WHERE s.deleted_at = 0
) drifted

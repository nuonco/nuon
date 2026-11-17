SELECT
    live.*
FROM (
    SELECT
        r.id,
        r.created_by_id,
        r.created_at,
        r.updated_at,
        r.deleted_at,
        r.org_id,
        sb.install_id,
        sb.id AS install_sandbox_id,
        r.install_workflow_id,
        r.role,
        r.planned_at,
        r.applied_at,
        r.planned_inputs_digest,
        r.applied_inputs_digest,
        r.run_type,
        r.status,
        r.status_description,
        r.status_v2,
        r.composite_error,
        r.app_sandbox_config_id,
        r.app_sandbox_build_id
    FROM
        install_sandboxes sb
        JOIN install_sandbox_runs r ON r.id = sb.app_config_ref ->> 'applied_config_by_id'
    WHERE
        sb.app_config_ref ->> 'applied_config_by_type' = 'install_sandbox_runs'
        AND COALESCE(sb.app_config_ref ->> 'applied_config_id', '') <> ''
        AND r.deleted_at = 0
        AND r.run_type IS DISTINCT FROM 'deprovision'
    UNION ALL
    (SELECT
        pick.id,
        pick.created_by_id,
        pick.created_at,
        pick.updated_at,
        pick.deleted_at,
        pick.org_id,
        pick.install_id,
        pick.install_sandbox_id,
        pick.install_workflow_id,
        pick.role,
        pick.planned_at,
        pick.applied_at,
        pick.planned_inputs_digest,
        pick.applied_inputs_digest,
        pick.run_type,
        pick.status,
        pick.status_description,
        pick.status_v2,
        pick.composite_error,
        pick.app_sandbox_config_id,
        pick.app_sandbox_build_id
    FROM (
        SELECT DISTINCT ON (sb.install_id, sb.id)
            r.id,
            r.created_by_id,
            r.created_at,
            r.updated_at,
            r.deleted_at,
            r.org_id,
            sb.install_id,
            sb.id AS install_sandbox_id,
            r.install_workflow_id,
            r.role,
            r.planned_at,
            r.applied_at,
            r.planned_inputs_digest,
            r.applied_inputs_digest,
            r.run_type,
            r.status,
            r.status_description,
            r.status_v2,
            r.composite_error,
            r.app_sandbox_config_id,
            r.app_sandbox_build_id
        FROM
            install_sandboxes sb
            JOIN install_sandbox_runs r ON r.install_sandbox_id = sb.id
        WHERE
            COALESCE(sb.app_config_ref ->> 'applied_config_by_id', '') = ''
            AND r.deleted_at = 0
            AND r.applied_at IS NOT NULL
        ORDER BY
            sb.install_id,
            sb.id,
            r.applied_at DESC,
            r.created_at DESC
    ) pick
    WHERE
        pick.run_type IS DISTINCT FROM 'deprovision')
) live

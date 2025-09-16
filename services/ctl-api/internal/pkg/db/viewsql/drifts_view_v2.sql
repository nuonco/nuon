-- Create view that shows unique drifted installs from both deploy and sandbox runs
WITH latest_drifted_deploys AS (
    SELECT
        id,
        install_component_id,
        ROW_NUMBER() OVER (PARTITION BY install_component_id ORDER BY created_at DESC) as rn
    FROM
        install_deploys
    WHERE
        status = 'drifted'
),
latest_drifted_sandbox_runs AS (
    SELECT
        id,
        install_sandbox_id,
        ROW_NUMBER() OVER (PARTITION BY install_sandbox_id ORDER BY created_at DESC) as rn
    FROM
        install_sandbox_runs
    WHERE
        status = 'drifted'
),
latest_deploy_status AS (
    SELECT
        install_component_id,
        status,
        ROW_NUMBER() OVER (PARTITION BY install_component_id ORDER BY created_at DESC) as rn
    FROM
        install_deploys
),
latest_sandbox_status AS (
    SELECT
        install_id,
        install_sandbox_id,
        status,
        ROW_NUMBER() OVER (PARTITION BY install_sandbox_id ORDER BY created_at DESC) as rn
    FROM
        install_sandbox_runs
)

SELECT
    'install_deploy' AS target_type,
    id.id AS target_id,
    id.install_workflow_id,
    NULL AS app_sandbox_config_id,
    id.component_build_id,
    ic.install_id,
    i.org_id,
    id.install_component_id,
    NULL AS install_sandbox_id,
    c.name AS component_name
FROM
    install_deploys id
JOIN
    install_components ic ON id.install_component_id = ic.id
JOIN
    installs i ON ic.install_id = i.id
JOIN
    components c ON ic.component_id = c.id
JOIN
    latest_drifted_deploys ldd ON id.id = ldd.id AND ldd.rn = 1
LEFT JOIN
    latest_deploy_status lds ON id.install_component_id = lds.install_component_id AND lds.rn = 1
WHERE
    id.status = 'drifted'
    AND (lds.status != 'no-drift' OR lds.status IS NULL)

UNION ALL

SELECT
    'install_sandbox_run' AS target_type,
    isr.id AS target_id,
    isr.install_workflow_id,
    isr.app_sandbox_config_id,
    NULL AS component_build_id,
    isr.install_id,
    i.org_id,
    NULL AS install_component_id,
    isr.install_sandbox_id,
    NULL AS component_name
FROM
    install_sandbox_runs isr
JOIN
    installs i ON isr.install_id = i.id
JOIN
    latest_drifted_sandbox_runs ldsr ON isr.id = ldsr.id AND ldsr.rn = 1
LEFT JOIN
    latest_sandbox_status lss ON isr.install_sandbox_id = lss.install_sandbox_id AND lss.rn = 1
WHERE
    isr.status = 'drifted'
    AND (lss.status != 'no-drift' OR lss.status IS NULL)

ORDER BY
    target_type, target_id;

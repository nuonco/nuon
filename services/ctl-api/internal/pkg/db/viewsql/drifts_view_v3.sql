-- Create view that shows unique drifted installs from both deploy and sandbox runs.
-- v3: replaces the fleet-wide latest_deploy_status / latest_sandbox_status window CTEs
-- with NOT EXISTS anti-joins so the "still drifted" guard no longer scans the entire
-- install_deploys / install_sandbox_runs tables on every query.
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
WHERE
    id.status = 'drifted'
    AND NOT EXISTS (
        SELECT 1
        FROM install_deploys d2
        WHERE d2.install_component_id = id.install_component_id
          AND d2.created_at > id.created_at
    )

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
WHERE
    isr.status = 'drifted'
    AND NOT EXISTS (
        SELECT 1
        FROM install_sandbox_runs sr2
        WHERE sr2.install_sandbox_id = isr.install_sandbox_id
          AND sr2.created_at > isr.created_at
    )

ORDER BY
    target_type, target_id;

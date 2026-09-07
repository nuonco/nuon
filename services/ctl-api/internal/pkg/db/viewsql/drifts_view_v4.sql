-- Create view that shows unique drifted installs from both deploy and sandbox runs.
-- v4: removes the latest_drifted_deploys / latest_drifted_sandbox_runs CTEs. The
-- ROW_NUMBER() window functions in those CTEs were redundant with the NOT EXISTS
-- anti-joins: if a row is drifted AND no newer row exists (any status), it is
-- automatically the latest drifted row. The CTEs forced materialization before the
-- caller's WHERE org_id/install_id filter could be pushed down, so the query scanned
-- all drifted deploys/sandbox runs across every org on every call. Without the CTEs,
-- PostgreSQL inlines the view and applies the caller's filter through the joins.
--
-- The NOT EXISTS tiebreaker (d2.created_at > id.created_at OR (d2.created_at = ... AND d2.id > ...))
-- preserves the CTE's one-row-per-component deduplication when two drifted rows share
-- the same created_at, without needing a window function.
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
WHERE
    id.status = 'drifted'
    AND NOT EXISTS (
        SELECT 1
        FROM install_deploys d2
        WHERE d2.install_component_id = id.install_component_id
          AND (
              d2.created_at > id.created_at
              OR (d2.created_at = id.created_at AND d2.id > id.id)
          )
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
WHERE
    isr.status = 'drifted'
    AND NOT EXISTS (
        SELECT 1
        FROM install_sandbox_runs sr2
        WHERE sr2.install_sandbox_id = isr.install_sandbox_id
          AND (
              sr2.created_at > isr.created_at
              OR (sr2.created_at = isr.created_at AND sr2.id > isr.id)
          )
    )

ORDER BY
    target_type, target_id;

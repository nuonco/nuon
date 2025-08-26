/* Load the most recent sandbox run */
WITH sandbox_runs_partitioned AS (
    SELECT
        sr.install_id,
        sr.status AS sandbox_status,
        sr.status_description AS sandbox_status_run,
        ROW_NUMBER() OVER (
            PARTITION BY sr.install_id
            ORDER BY
                sr.created_at DESC
        ) AS rn
    FROM
        install_sandbox_runs sr
    WHERE
        sr.deleted_at = 0
        AND sr.status != 'drift-detected'
        AND sr.status != 'auto-skipped'
        AND sr.status != 'no-drift'
),
/* filter down to the most recent install sandbox runs */
latest_sandbox_runs AS (
    SELECT
        srp.*
    FROM
        sandbox_runs_partitioned srp
    WHERE
        srp.rn = 1
),
/* Build a mapping of the components and statuses directly from install_components */
aggregated_component_statuses AS (
    SELECT
        ic.install_id,
        hstore(array_agg(ic.component_id), array_agg(ic.status)) AS component_statuses
    FROM
        install_components ic
    WHERE
        ic.deleted_at = 0
    GROUP BY
        ic.install_id
)
/* Build the final installs table */
SELECT
    i.*,
    sandbox_status,
    sandbox_status_run,
    component_statuses,
    row_number() OVER (
        PARTITION BY app_id
        ORDER BY
            created_at
    ) AS install_number
FROM
    installs i FULL
    OUTER JOIN latest_sandbox_runs lsr ON i.id = lsr.install_id FULL
    OUTER JOIN aggregated_component_statuses acs ON i.id = acs.install_id

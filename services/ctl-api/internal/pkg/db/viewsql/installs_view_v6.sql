/* Build a mapping of the components and statuses directly from install_components */
WITH aggregated_component_statuses AS (
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
    is_data.status AS sandbox_status,
    is_data.status_description AS sandbox_status_run,
    component_statuses,
    row_number() OVER (
        PARTITION BY app_id
        ORDER BY
            is_data.created_at
    ) AS install_number
FROM
    installs i
    LEFT JOIN install_sandboxes is_data ON i.id = is_data.install_id AND is_data.deleted_at = 0
    FULL OUTER JOIN aggregated_component_statuses acs ON i.id = acs.install_id

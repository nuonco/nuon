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
/* Build the final installs view.
 *
 * Compared to v7, this view replaces the FULL OUTER JOIN against the
 * component statuses CTE with a LEFT JOIN — installs are always the
 * driving table, so the FULL OUTER added no rows but blocked predicate
 * pushdown of WHERE id = $1 through to the underlying tables.
 *
 * The install_number column is computed via a LATERAL subquery instead
 * of a window function over the entire result set. The window function
 * forced a sort of every install in the database to answer a single
 * by-id lookup; the LATERAL scopes the count to the install's own
 * app_id partition, which is small.
 */
SELECT
    i.*,
    is_data.status AS sandbox_status,
    is_data.status_description AS sandbox_status_run,
    acs.component_statuses,
    rn.install_number
FROM
    installs i
    LEFT JOIN install_sandboxes is_data
        ON i.id = is_data.install_id AND is_data.deleted_at = 0
    LEFT JOIN aggregated_component_statuses acs
        ON i.id = acs.install_id
    LEFT JOIN LATERAL (
        SELECT count(*) + 1 AS install_number
        FROM installs i2
        LEFT JOIN install_sandboxes is2
            ON is2.install_id = i2.id AND is2.deleted_at = 0
        WHERE i2.app_id = i.app_id
          AND i2.deleted_at = 0
          AND (
              COALESCE(is2.created_at, 'infinity'::timestamptz),
              i2.id
          ) < (
              COALESCE(is_data.created_at, 'infinity'::timestamptz),
              i.id
          )
    ) rn ON true

SELECT
    i.*,
    is_data.status AS sandbox_status,
    is_data.status_description AS sandbox_status_run,
    (
        SELECT hstore(array_agg(ic.component_id), array_agg(ic.status))
        FROM install_components ic
        WHERE ic.deleted_at = 0 AND ic.install_id = i.id
    ) AS component_statuses,
    (
        SELECT n FROM (
            SELECT i2.id AS id,
                   row_number() OVER (ORDER BY is2.created_at) AS n
            FROM installs i2
            LEFT JOIN install_sandboxes is2
              ON i2.id = is2.install_id AND is2.deleted_at = 0
            WHERE i2.app_id = i.app_id
        ) sub
        WHERE sub.id = i.id
    ) AS install_number
FROM installs i
LEFT JOIN install_sandboxes is_data
  ON i.id = is_data.install_id AND is_data.deleted_at = 0

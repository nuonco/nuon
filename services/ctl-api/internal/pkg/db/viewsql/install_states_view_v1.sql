/* Build the final install state table */
SELECT
    i.*,
    row_number() OVER (
        PARTITION BY install_id
        ORDER BY
            created_at
    ) AS version
FROM
    install_states i
WHERE
    i.revision = 0
    OR EXISTS (
        SELECT 1
        FROM install_state_publications p
        WHERE p.state_id = i.id
    )

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

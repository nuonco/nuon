SELECT
    (
        SELECT n FROM (
            SELECT ac2.id AS id,
                   row_number() OVER (ORDER BY ac2.created_at) AS n
            FROM app_configs ac2
            WHERE ac2.app_id = ac.app_id
        ) sub
        WHERE sub.id = ac.id
    ) AS version,
    ac.*
FROM app_configs ac

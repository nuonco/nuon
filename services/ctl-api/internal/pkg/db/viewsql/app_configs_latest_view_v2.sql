SELECT a.*
FROM app_configs a
INNER JOIN (
    SELECT app_id, MAX(version) AS max_version
    FROM app_configs
    WHERE deleted_at = 0
    GROUP BY app_id
) latest ON a.app_id = latest.app_id AND a.version = latest.max_version
WHERE a.deleted_at = 0

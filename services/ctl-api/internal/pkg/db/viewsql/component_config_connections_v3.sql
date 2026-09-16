SELECT ccc.*,
  (SELECT count(*) FROM component_config_connections c2
    WHERE c2.component_id = ccc.component_id AND (c2.created_at, c2.id) <= (ccc.created_at, ccc.id)) AS version,
  (SELECT count(*) FROM component_config_connections c2
    WHERE c2.component_id = ccc.component_id AND (c2.created_at, c2.id) >= (ccc.created_at, ccc.id)) AS execution_number,
  ac.version AS app_config_version
FROM component_config_connections ccc
JOIN (SELECT id, row_number() OVER (PARTITION BY app_id ORDER BY created_at) AS version FROM app_configs) ac
  ON ac.id = ccc.app_config_id

SELECT ccc.*,
  (SELECT count(*) FROM component_config_connections c2
    WHERE c2.component_id = ccc.component_id AND (c2.created_at, c2.id) <= (ccc.created_at, ccc.id)) AS version,
  (SELECT count(*) FROM component_config_connections c2
    WHERE c2.component_id = ccc.component_id AND (c2.created_at, c2.id) >= (ccc.created_at, ccc.id)) AS execution_number,
  (SELECT count(*) FROM app_configs ac2
    WHERE ac2.app_id = ac.app_id AND (ac2.created_at, ac2.id) <= (ac.created_at, ac.id)) AS app_config_version
FROM component_config_connections ccc
JOIN app_configs ac ON ac.id = ccc.app_config_id

SELECT
  ccc.*,
  ac.version as app_config_version
FROM component_config_connections ccc
JOIN
  app_configs ac on ac.id = ccc.app_config_id
ORDER BY ccc.version DESC

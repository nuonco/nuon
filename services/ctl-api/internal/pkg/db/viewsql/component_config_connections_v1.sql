WITH component_config_connections_with_count AS (
  SELECT
       rje.*,

       ROW_NUMBER() OVER (
	  PARTITION BY component_id
	  ORDER BY
	    created_at
       ) AS version,

       ROW_NUMBER() OVER (PARTITION BY rje.component_id ORDER BY rje.created_at DESC) as execution_number
  FROM
     component_config_connections rje
)

SELECT
  ccc.*,
  ac.version as app_config_version
FROM component_config_connections_with_count ccc
JOIN 
  app_configs_view_v2 ac on ac.id = ccc.app_config_id
ORDER BY ccc.version DESC

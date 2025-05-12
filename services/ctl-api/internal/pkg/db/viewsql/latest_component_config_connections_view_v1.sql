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
	rje.*
FROM
	component_config_connections_with_count rje
WHERE
	rje.execution_number = 1

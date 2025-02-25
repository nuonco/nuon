  /* Load the count of configs */
  /* Build a mapping of the components and statuses */
  WITH action_workflow_configs_aggregated AS (
      SELECT 
    	awc.action_workflow_id,
      count(*) AS config_count
      FROM action_workflow_configs as awc
      GROUP BY (action_workflow_id)
  )

  /* Build the final installs table */
  SELECT
      aw.*,
      awca.config_count
  FROM 
      action_workflows aw
  JOIN 
      action_workflow_configs_aggregated awca
  ON 
      aw.id = awca.action_workflow_id

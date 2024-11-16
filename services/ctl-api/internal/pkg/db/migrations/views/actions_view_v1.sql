CREATE OR REPLACE VIEW actions_view_v1 AS
  /* Load the count of configs */
  /* Build a mapping of the components and statuses */
  WITH action_configs_aggregated AS (
      SELECT 
    	ac.action_id,
      count(*) AS config_count
      FROM action_configs as ac
      GROUP BY (action_id)
  )

  /* Build the final installs table */
  SELECT
      a.*,
      ac.config_count
  FROM 
      actions a
  JOIN 
      action_configs_aggregated ac
  ON 
      a.id = ac.action_id

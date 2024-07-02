CREATE OR REPLACE VIEW installs_view_v2 AS
  /* Load the most recent sandbox run */
  WITH sandbox_runs_partitioned AS (
      SELECT 
        sr.install_id,
        sr.status AS sandbox_status,
        sr.status_description AS sandbox_status_run,
    	ROW_NUMBER() OVER (PARTITION BY sr.install_id ORDER BY sr.created_at DESC) AS rn
      FROM 
    install_sandbox_runs sr
  ),

  /* filter down to the most recent install sandbox runs */
  latest_sandbox_runs AS (
      SELECT 
        srp.*
      FROM sandbox_runs_partitioned srp
      WHERE srp.rn = 1
  ),

  /* Load the most recent deploys for each install component */
  install_deploys_partitioned AS (
      SELECT 
        ic.install_id,
        ic.component_id,
        id.status,
    ROW_NUMBER() OVER (PARTITION BY ic.install_id, ic.component_id ORDER BY id.created_at DESC) AS rn
      FROM 
    install_deploys id
      JOIN
    install_components ic on ic.id=id.install_component_id
  ),

  /* Build a mapping of the components and statuses */
  aggregated_latest_install_deploys AS (
      SELECT 
    	ld.install_id,
    	hstore(array_agg(ld.component_id), array_agg(ld.status)) as component_statuses
      FROM install_deploys_partitioned as ld
      WHERE 
        ld.rn = 1
      GROUP BY (install_id, status)
  )

  /* Build the final installs table */
  SELECT
      i.*,
      sandbox_status,
      sandbox_status_run,
      component_statuses,
      row_number() OVER (PARTITION BY app_id ORDER BY created_at) AS install_number
  FROM 
      installs i
  FULL OUTER JOIN 
      latest_sandbox_runs lsr
  ON 
      i.id = lsr.install_id
  FULL OUTER JOIN 
      aggregated_latest_install_deploys ld
  ON  i.id = ld.install_id

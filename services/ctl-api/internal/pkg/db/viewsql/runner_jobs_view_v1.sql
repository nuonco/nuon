  /* Runner job execution counts */
  WITH runner_job_execution_counts AS (
    SELECT
        rje.runner_job_id,
        count(*) as execution_count
    FROM 
        runner_job_executions rje
    GROUP BY runner_job_id  
  ),

  runner_job_executions AS (
    SELECT
       rje.*,
       ROW_NUMBER() OVER (PARTITION BY rje.runner_job_id ORDER BY rje.created_at) as execution_number
    FROM
       runner_job_executions rje

  ),

  runner_job_executions_latest AS (
    SELECT
       rje.*
    FROM
       runner_job_executions rje
    WHERE
      rje.execution_number = 1
  ),

  runner_job_executions_latest_with_outputs AS (
    SELECT
      rjel.*,
      rjeo.outputs
    FROM
      runner_job_executions_latest rjel
    JOIN
      runner_job_execution_outputs rjeo
    ON
      rjel.id = rjeo.runner_job_execution_id
  )

  SELECT 
    rj.*,
    rjec.execution_count,
    rjel.id AS final_runner_job_execution_id,
    rjel.outputs AS outputs
  FROM    
    runner_jobs AS rj
  FULL OUTER JOIN
    runner_job_execution_counts AS rjec
  ON 
    rjec.runner_job_id = rj.id
  FULL OUTER JOIN
    runner_job_executions_latest_with_outputs AS rjel
  ON
    rjel.runner_job_id = rj.id

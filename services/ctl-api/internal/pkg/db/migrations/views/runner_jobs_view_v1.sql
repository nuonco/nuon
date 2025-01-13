DROP VIEW IF EXISTS runner_jobs_view_v1;

CREATE OR REPLACE VIEW runner_jobs_view_v1 AS
  /* Runner job execution counts */
  WITH runner_job_execution_counts AS (
    SELECT
        rje.runner_job_id,
        count(*) as execution_count
    FROM 
        runner_job_executions rje
    GROUP BY runner_job_id  
  )

  SELECT 
    rj.*,
    rjec.execution_count
  FROM    
    runner_jobs AS rj
  FULL OUTER JOIN
    runner_job_execution_counts AS rjec
  ON 
    rjec.runner_job_id = rj.id

SELECT
    rj.*,
    -- Execution count subquery with NULL if 0
    NULLIF(
        (
            SELECT
                count(*)
            FROM
                public.runner_job_executions rje
            WHERE
                rje.runner_job_id = rj.id
        ),
        0
    ) AS execution_count,
    -- Final execution ID subquery
    (
        SELECT
            rjeo.runner_job_execution_id
        FROM
            runner_job_execution_outputs rjeo
            JOIN (
                SELECT
                    rje.id
                FROM
                    public.runner_job_executions rje
                WHERE
                    rje.runner_job_id = rj.id
                ORDER BY
                    rje.created_at
                LIMIT
                    1
            ) first_exec ON first_exec.id = rjeo.runner_job_execution_id
        LIMIT
            1
    ) AS final_runner_job_execution_id,
    -- Outputs blob metadata subquery. Carries the S3 pointer rather than the
    -- payload, which lives in the blob and cannot be read from SQL.
    (
        SELECT
            rjeo.outputs_blob
        FROM
            runner_job_execution_outputs rjeo
            JOIN (
                SELECT
                    rje.id
                FROM
                    public.runner_job_executions rje
                WHERE
                    rje.runner_job_id = rj.id
                ORDER BY
                    rje.created_at
                LIMIT
                    1
            ) first_exec ON first_exec.id = rjeo.runner_job_execution_id
        LIMIT
            1
    ) AS outputs_blob
FROM
    runner_jobs rj;

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
    -- Outputs subquery
    (
        SELECT
            rjeo.outputs
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
    ) AS outputs
FROM
    runner_jobs rj;

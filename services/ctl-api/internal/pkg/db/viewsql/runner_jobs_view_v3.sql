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
    -- Final execution ID subquery: the latest execution with outputs, finished ones first
    (
        SELECT
            rjeo.runner_job_execution_id
        FROM
            runner_job_execution_outputs rjeo
            JOIN public.runner_job_executions rje ON rje.id = rjeo.runner_job_execution_id
        WHERE
            rje.runner_job_id = rj.id
        ORDER BY
            rje.status = 'finished' DESC,
            rje.created_at DESC
        LIMIT
            1
    ) AS final_runner_job_execution_id,
    -- Outputs subquery
    (
        SELECT
            rjeo.outputs
        FROM
            runner_job_execution_outputs rjeo
            JOIN public.runner_job_executions rje ON rje.id = rjeo.runner_job_execution_id
        WHERE
            rje.runner_job_id = rj.id
        ORDER BY
            rje.status = 'finished' DESC,
            rje.created_at DESC
        LIMIT
            1
    ) AS outputs
FROM
    runner_jobs rj;

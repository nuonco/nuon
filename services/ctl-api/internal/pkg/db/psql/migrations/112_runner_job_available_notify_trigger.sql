-- Emit a Postgres NOTIFY whenever a runner job becomes available so the
-- long-poll job-pickup endpoint (TailRunnerJobs) can wake parked runners in
-- ~ms instead of waiting out its poll backstop. The notification carries no
-- correctness: the endpoint always re-queries Postgres on wake, and keeps a
-- sparse poll as a backstop, so a dropped notify (listener reconnecting, pod
-- restart, RDS failover) only costs latency, never a stranded job.
--
-- The function fires for both INSERT (job born available) and the
-- queued -> available UPDATE. Logic lives in the body (not a WHEN clause) so a
-- single trigger can cover INSERT OR UPDATE without referencing OLD on INSERT.

CREATE OR REPLACE FUNCTION notify_runner_job_available() RETURNS trigger AS $$
BEGIN
  IF NEW.status = 'available'
     AND (TG_OP = 'INSERT' OR OLD.status IS DISTINCT FROM NEW.status) THEN
    PERFORM pg_notify(
      'runner_job_available_v1',
      json_build_object(
        'runner_id', NEW.runner_id,
        'job_id', NEW.id,
        'group', NEW."group"
      )::text
    );
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS runner_job_available_notify ON runner_jobs;

CREATE TRIGGER runner_job_available_notify
AFTER INSERT OR UPDATE OF status ON runner_jobs
FOR EACH ROW
EXECUTE FUNCTION notify_runner_job_available();

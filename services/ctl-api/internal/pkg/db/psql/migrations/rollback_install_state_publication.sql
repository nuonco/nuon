\set ON_ERROR_STOP on
\if :{?install_state_writers_stopped}
\else
\set install_state_writers_stopped off
\endif
\if :install_state_writers_stopped
\else
DO $$ BEGIN
  RAISE EXCEPTION 'Requires -v install_state_writers_stopped=on after draining all ctl-api processes';
END; $$;
\endif

SET lock_timeout = '2s';

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM (
      SELECT DISTINCT ON (s.install_id) s.state
      FROM install_states s
      JOIN installs i ON i.id = s.install_id AND i.deleted_at = 0
      WHERE s.deleted_at = 0
      ORDER BY s.install_id, s.created_at DESC, s.id DESC
    ) latest
    WHERE NULLIF(latest.state, 'null'::jsonb) IS NULL
  ) THEN
    RAISE EXCEPTION 'Restore inline JSON for current blob-only snapshots before rollback';
  END IF;
END;
$$;

SELECT format('DROP TRIGGER %I ON %I.%I;', t.tgname, n.nspname, c.relname)
FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_proc p ON p.oid = t.tgfoid
WHERE n.nspname = 'public' AND NOT t.tgisinternal
  AND starts_with(p.proname, 'install_state_invalidate_')
ORDER BY c.relname, t.tgname
\gexec

BEGIN;
UPDATE install_states
SET stale_at = now(),
    stale_partials = '["org","app","domain","runner","cloud","actions","inputs","components","sandbox","stack","secrets"]'::jsonb,
    state_blob = CASE WHEN NULLIF(state, 'null'::jsonb) IS NOT NULL THEN NULL ELSE state_blob END
WHERE deleted_at = 0 AND (NOT archived OR NULLIF(state, 'null'::jsonb) IS NOT NULL);

DELETE FROM install_state_heads;

DELETE FROM migrations
WHERE name IN ('143-install-state-invalidation-triggers', '144-install-state-invalidation-functions', '146-install-state-heads')
   OR starts_with(name, '145-install-state-invalidation-');
COMMIT;

SELECT format('DROP FUNCTION %s;', p.oid::regprocedure)
FROM pg_proc p
JOIN pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname = 'public'
  AND (starts_with(p.proname, 'install_state_invalidate_')
       OR p.proname IN ('invalidate_install_state', 'invalidate_install_state_targets'))
ORDER BY p.proname
\gexec

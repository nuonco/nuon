# Drifts view v6 rollback

`drifted_objects_view_v6` reads the install deploy, sandbox run, and stack version
live views and the selected-build view. All four are reapplied on every startup
migration: the migrator drops every registered reapplied view in one statement,
without `CASCADE`, then recreates them.

A release older than v6 does not register it, so its startup migration cannot drop
the live views while v6 still depends on them, and fails. API pods from either
release keep working: the old ones read `drifted_objects_view_v5`, which v6 leaves
unchanged.

## Roll back

1. Stop the new release's API processes. Their install overview reads v6.
2. Verify the intended database connection, then run:

   ```sh
   psql "$DATABASE_URL" \
     -f services/ctl-api/internal/pkg/db/viewsql/rollback_drifts_view_v6.sql
   ```

   It drops only v6, without `CASCADE`, under a two-second lock timeout. The view
   holds no data, and the script is safe to rerun.
3. Run the old release's startup migrations, then start its processes.

Rolling forward again recreates v6 on the next startup migration.

## Retiring v5

Removing v5 from the registry does not drop it, and it would then block the new
release's migrator in the same way. Drop it in a migration instead.

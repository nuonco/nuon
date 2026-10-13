CREATE OR REPLACE FUNCTION invalidate_install_state_targets(target_install_ids text[], target_partials text[]) RETURNS void AS $$
DECLARE
  known CONSTANT text[] := ARRAY['org', 'app', 'domain', 'runner', 'cloud', 'actions', 'inputs', 'components', 'sandbox', 'stack', 'secrets'];
  invalid text;
BEGIN
  IF cardinality(target_install_ids) IS DISTINCT FROM cardinality(target_partials) THEN
    RAISE EXCEPTION 'install state invalidation targets and partials must have the same length';
  END IF;
  IF coalesce(cardinality(target_install_ids), 0) = 0 THEN
    RETURN;
  END IF;

  SELECT t.partial INTO invalid
  FROM unnest(target_partials) AS t(partial)
  WHERE t.partial IS NULL OR NOT t.partial = ANY(known)
  LIMIT 1;
  IF FOUND THEN
    RAISE EXCEPTION 'unknown install state partial: %', coalesce(invalid, '<null>');
  END IF;

  PERFORM 1
  FROM install_state_heads h
  WHERE h.install_id = ANY(target_install_ids)
  ORDER BY h.install_id
  FOR NO KEY UPDATE OF h;

  INSERT INTO install_state_heads AS h (install_id, org_id, revision, current_state_id, stale_at, stale_partials)
  SELECT i.id, i.org_id, 1, '', now(), t.partials
  FROM (
    SELECT d.install_id, jsonb_agg(d.partial ORDER BY array_position(known, d.partial)) AS partials
    FROM (
      SELECT DISTINCT u.install_id, u.partial
      FROM unnest(target_install_ids, target_partials) AS u(install_id, partial)
      WHERE u.install_id IS NOT NULL
    ) d
    GROUP BY d.install_id
  ) t
  JOIN installs i ON i.id = t.install_id AND i.deleted_at = 0
  ORDER BY i.id
  ON CONFLICT (install_id) DO UPDATE SET
    org_id = EXCLUDED.org_id,
    revision = h.revision + 1,
    stale_at = CASE
      WHEN h.stale_at IS NULL OR h.stale_partials IS NULL OR h.stale_partials = '[]'::jsonb THEN EXCLUDED.stale_at
      ELSE h.stale_at
    END,
    stale_partials = (
      SELECT coalesce(jsonb_agg(m.partial ORDER BY array_position(known, m.partial), m.partial), '[]'::jsonb)
      FROM (
        SELECT jsonb_array_elements_text(CASE WHEN jsonb_typeof(h.stale_partials) = 'array' THEN h.stale_partials ELSE '[]'::jsonb END)
        UNION
        SELECT jsonb_array_elements_text(EXCLUDED.stale_partials)
      ) AS m(partial)
    );
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION invalidate_install_state(install_ids text[], partials jsonb) RETURNS void AS $$
BEGIN
  IF partials IS NULL OR jsonb_typeof(partials) <> 'array' THEN
    RAISE EXCEPTION 'install state partials must be a json array';
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(i.install_id), array_agg(p.partial))
  FROM unnest(install_ids) AS i(install_id)
  CROSS JOIN jsonb_array_elements_text(partials) AS p(partial);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_runner_job_targets(jobs runner_jobs[], outputs_only boolean)
RETURNS TABLE(install_id text, partial text) AS $$
  SELECT ic.install_id, 'components'
  FROM unnest(jobs) j
  JOIN install_deploys d ON d.id = j.owner_id
  JOIN install_components ic ON ic.id = d.install_component_id
  WHERE j.owner_type = 'install_deploys'
    AND j.operation IS DISTINCT FROM 'create-apply-plan'
    AND d.status IN ('active', 'inactive')
  UNION ALL
  SELECT r.install_id, p.partial
  FROM unnest(jobs) j
  JOIN install_sandbox_runs r ON r.id = j.owner_id
  CROSS JOIN unnest(ARRAY['sandbox', 'domain']) AS p(partial)
  WHERE j.owner_type = 'install_sandbox_runs'
    AND r.status = 'active'
  UNION ALL
  SELECT r.install_id, 'actions'
  FROM unnest(jobs) j
  JOIN install_action_workflow_runs r ON r.id = j.owner_id
  WHERE j.owner_type = 'install_action_workflow_runs'
    AND r.status = 'finished'
    AND (NOT outputs_only OR j.type = 'actions-workflow')
  UNION ALL
  SELECT r.install_id, 'actions'
  FROM unnest(jobs) j
  JOIN install_action_workflow_run_steps s ON s.id = j.owner_id
  JOIN install_action_workflow_runs r ON r.id = s.install_action_workflow_run_id
  WHERE j.owner_type = 'install_action_workflow_run_steps'
    AND r.status = 'finished'
  UNION ALL
  SELECT g.owner_id, 'secrets'
  FROM unnest(jobs) j
  JOIN runners ru ON ru.id = j.runner_id
  JOIN runner_groups g ON g.id = ru.runner_group_id
  WHERE j.type = 'sandbox-sync-secrets'
    AND g.owner_type = 'installs'
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION install_state_invalidate_installs() RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    PERFORM invalidate_install_state(array_agg(n.id ORDER BY n.id), '["org", "app", "domain", "runner", "cloud", "actions", "inputs", "components", "sandbox", "stack", "secrets"]'::jsonb)
    FROM new_rows n;
  ELSIF TG_OP = 'UPDATE' THEN
    PERFORM invalidate_install_state_targets(array_agg(n.id), array_agg(x.partial))
    FROM old_rows o
    JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (
      SELECT 'org' WHERE o.org_id IS DISTINCT FROM n.org_id
      UNION ALL
      SELECT 'app' WHERE o.app_id IS DISTINCT FROM n.app_id
      UNION ALL
      SELECT 'inputs' WHERE o.app_config_id IS DISTINCT FROM n.app_config_id
      UNION ALL
      SELECT 'cloud' WHERE o.name IS DISTINCT FROM n.name
      UNION ALL
      SELECT a.partial
      FROM unnest(ARRAY['org', 'app', 'domain', 'runner', 'cloud', 'actions', 'inputs', 'components', 'sandbox', 'stack', 'secrets']) AS a(partial)
      WHERE o.deleted_at IS DISTINCT FROM n.deleted_at
    ) AS x(partial);
  ELSE
    DELETE FROM install_state_publications p USING old_rows o WHERE p.install_id = o.id;
    DELETE FROM install_state_heads h USING old_rows o WHERE h.install_id = o.id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_orgs() RETURNS trigger AS $$
DECLARE
  rs orgs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::orgs), (n::orgs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.name, o.status, o.deleted_at) IS DISTINCT FROM ROW(n.name, n.status, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(i.id), array_agg('org'::text))
  FROM unnest(rs) r
  JOIN installs i ON i.org_id = r.id;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_apps() RETURNS trigger AS $$
DECLARE
  rs apps[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::apps), (n::apps)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.name, o.status, o.deleted_at) IS DISTINCT FROM ROW(n.name, n.status, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(i.id), array_agg('app'::text))
  FROM unnest(rs) r
  JOIN installs i ON i.app_id = r.id;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_app_secrets() RETURNS trigger AS $$
DECLARE
  rs app_secrets[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::app_secrets), (n::app_secrets)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.app_id, o.name, o.value, o.deleted_at) IS DISTINCT FROM ROW(n.app_id, n.name, n.value, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(i.id), array_agg('app'::text))
  FROM unnest(rs) r
  JOIN installs i ON i.app_id = r.app_id;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_runner_groups() RETURNS trigger AS $$
DECLARE
  rs runner_groups[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::runner_groups), (n::runner_groups)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.owner_id, o.owner_type, o.deleted_at) IS DISTINCT FROM ROW(n.owner_id, n.owner_type, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.owner_id), array_agg(p.partial))
  FROM unnest(rs) r
  CROSS JOIN unnest(ARRAY['runner', 'secrets']) AS p(partial)
  WHERE r.owner_type = 'installs';
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_runners() RETURNS trigger AS $$
DECLARE
  rs runners[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::runners), (n::runners)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.runner_group_id, o.status, o.deleted_at) IS DISTINCT FROM ROW(n.runner_group_id, n.status, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(g.owner_id), array_agg(p.partial))
  FROM unnest(rs) r
  JOIN runner_groups g ON g.id = r.runner_group_id
  CROSS JOIN unnest(ARRAY['runner', 'secrets']) AS p(partial)
  WHERE g.owner_type = 'installs';
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_aws_accounts() RETURNS trigger AS $$
DECLARE
  rs aws_accounts[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::aws_accounts), (n::aws_accounts)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.region, o.deleted_at) IS DISTINCT FROM ROW(n.install_id, n.region, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg('cloud'::text))
  FROM unnest(rs) r;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_azure_accounts() RETURNS trigger AS $$
DECLARE
  rs azure_accounts[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::azure_accounts), (n::azure_accounts)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.location, o.deleted_at) IS DISTINCT FROM ROW(n.install_id, n.location, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg('cloud'::text))
  FROM unnest(rs) r;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_gcp_accounts() RETURNS trigger AS $$
DECLARE
  rs gcp_accounts[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::gcp_accounts), (n::gcp_accounts)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.project_id, o.region, o.deleted_at) IS DISTINCT FROM ROW(n.install_id, n.project_id, n.region, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg('cloud'::text))
  FROM unnest(rs) r;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_app_input_configs() RETURNS trigger AS $$
DECLARE
  rs app_input_configs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::app_input_configs), (n::app_input_configs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.app_config_id, o.deleted_at) IS DISTINCT FROM ROW(n.app_config_id, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(i.id), array_agg('inputs'::text))
  FROM unnest(rs) r
  JOIN installs i ON i.app_config_id = r.app_config_id;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_app_inputs() RETURNS trigger AS $$
DECLARE
  rs app_inputs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::app_inputs), (n::app_inputs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.app_input_config_id, o.name, o."default", o.deleted_at) IS DISTINCT FROM ROW(n.app_input_config_id, n.name, n."default", n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(i.id), array_agg('inputs'::text))
  FROM unnest(rs) r
  JOIN app_input_configs c ON c.id = r.app_input_config_id
  JOIN installs i ON i.app_config_id = c.app_config_id;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_inputs() RETURNS trigger AS $$
DECLARE
  rs install_inputs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_inputs), (n::install_inputs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.values, o.created_at, o.deleted_at) IS DISTINCT FROM ROW(n.install_id, n.values, n.created_at, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg('inputs'::text))
  FROM unnest(rs) r;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_sandbox_runs() RETURNS trigger AS $$
DECLARE
  rs install_sandbox_runs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_sandbox_runs), (n::install_sandbox_runs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.status, o.status_v2->>'status', o.created_at, o.app_sandbox_config_id, o.deleted_at)
          IS DISTINCT FROM ROW(n.install_id, n.status, n.status_v2->>'status', n.created_at, n.app_sandbox_config_id, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg(p.partial))
  FROM unnest(rs) r
  CROSS JOIN unnest(ARRAY['sandbox', 'domain']) AS p(partial)
  WHERE r.status = 'active';
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_app_sandbox_configs() RETURNS trigger AS $$
DECLARE
  rs app_sandbox_configs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::app_sandbox_configs), (n::app_sandbox_configs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL OR o.deleted_at IS DISTINCT FROM n.deleted_at);
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(s.install_id), array_agg('sandbox'::text))
  FROM unnest(rs) r
  JOIN install_sandbox_runs s ON s.app_sandbox_config_id = r.id
  WHERE s.status = 'active';
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_public_git_vcs_configs() RETURNS trigger AS $$
DECLARE
  rs public_git_vcs_configs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::public_git_vcs_configs), (n::public_git_vcs_configs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.component_config_id, o.component_config_type, o.directory, o.branch, o.deleted_at)
          IS DISTINCT FROM ROW(n.component_config_id, n.component_config_type, n.directory, n.branch, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(s.install_id), array_agg('sandbox'::text))
  FROM unnest(rs) r
  JOIN install_sandbox_runs s ON s.app_sandbox_config_id = r.component_config_id
  WHERE r.component_config_type = 'app_sandbox_configs'
    AND s.status = 'active';
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_connected_github_vcs_configs() RETURNS trigger AS $$
DECLARE
  rs connected_github_vcs_configs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::connected_github_vcs_configs), (n::connected_github_vcs_configs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.component_config_id, o.component_config_type, o.directory, o.branch, o.deleted_at)
          IS DISTINCT FROM ROW(n.component_config_id, n.component_config_type, n.directory, n.branch, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(s.install_id), array_agg('sandbox'::text))
  FROM unnest(rs) r
  JOIN install_sandbox_runs s ON s.app_sandbox_config_id = r.component_config_id
  WHERE r.component_config_type = 'app_sandbox_configs'
    AND s.status = 'active';
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_stacks() RETURNS trigger AS $$
DECLARE
  rs install_stacks[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_stacks), (n::install_stacks)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.deleted_at) IS DISTINCT FROM ROW(n.install_id, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg('stack'::text))
  FROM unnest(rs) r;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_stack_versions() RETURNS trigger AS $$
DECLARE
  rs install_stack_versions[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_stack_versions), (n::install_stack_versions)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_stack_id, o.install_id, o.created_at, o.quick_link_url, o.template_url, o.contents, o.checksum, o.status->>'status', o.deleted_at)
          IS DISTINCT FROM ROW(n.install_stack_id, n.install_id, n.created_at, n.quick_link_url, n.template_url, n.contents, n.checksum, n.status->>'status', n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(t.install_id), array_agg('stack'::text))
  FROM (
    SELECT s.install_id
    FROM unnest(rs) r
    JOIN install_stacks s ON s.id = r.install_stack_id
    UNION
    SELECT r.install_id
    FROM unnest(rs) r
  ) t;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_stack_outputs() RETURNS trigger AS $$
DECLARE
  rs install_stack_outputs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_stack_outputs), (n::install_stack_outputs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_stack_id, o.data, o.deleted_at) IS DISTINCT FROM ROW(n.install_stack_id, n.data, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(s.install_id), array_agg('stack'::text))
  FROM unnest(rs) r
  JOIN install_stacks s ON s.id = r.install_stack_id;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_components() RETURNS trigger AS $$
DECLARE
  rs install_components[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_components), (n::install_components)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.component_id, o.deleted_at) IS DISTINCT FROM ROW(n.install_id, n.component_id, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg('components'::text))
  FROM unnest(rs) r;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_components() RETURNS trigger AS $$
DECLARE
  rs components[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::components), (n::components)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.name, o.deleted_at) IS DISTINCT FROM ROW(n.name, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(ic.install_id), array_agg('components'::text))
  FROM unnest(rs) r
  JOIN install_components ic ON ic.component_id = r.id;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_deploys() RETURNS trigger AS $$
DECLARE
  rs install_deploys[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_deploys), (n::install_deploys)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_component_id, o.status, o.status_v2->>'status', o.component_build_id, o.created_at, o.deleted_at)
          IS DISTINCT FROM ROW(n.install_component_id, n.status, n.status_v2->>'status', n.component_build_id, n.created_at, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(ic.install_id), array_agg('components'::text))
  FROM unnest(rs) r
  JOIN install_components ic ON ic.id = r.install_component_id
  WHERE r.status IN ('active', 'inactive');
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_action_workflows() RETURNS trigger AS $$
DECLARE
  rs install_action_workflows[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_action_workflows), (n::install_action_workflows)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.action_workflow_id, o.deleted_at) IS DISTINCT FROM ROW(n.install_id, n.action_workflow_id, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg('actions'::text))
  FROM unnest(rs) r;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_action_workflows() RETURNS trigger AS $$
DECLARE
  rs action_workflows[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::action_workflows), (n::action_workflows)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.name, o.deleted_at) IS DISTINCT FROM ROW(n.name, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(w.install_id), array_agg('actions'::text))
  FROM unnest(rs) r
  JOIN install_action_workflows w ON w.action_workflow_id = r.id;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_action_workflow_runs() RETURNS trigger AS $$
DECLARE
  rs install_action_workflow_runs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_action_workflow_runs), (n::install_action_workflow_runs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_id, o.install_action_workflow_id, o.status, o.status_v2->>'status', o.created_at, o.deleted_at)
          IS DISTINCT FROM ROW(n.install_id, n.install_action_workflow_id, n.status, n.status_v2->>'status', n.created_at, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(r.install_id), array_agg('actions'::text))
  FROM unnest(rs) r
  WHERE r.status = 'finished';
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_install_action_workflow_run_steps() RETURNS trigger AS $$
DECLARE
  rs install_action_workflow_run_steps[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::install_action_workflow_run_steps), (n::install_action_workflow_run_steps)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.install_action_workflow_run_id, o.deleted_at) IS DISTINCT FROM ROW(n.install_action_workflow_run_id, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(w.install_id), array_agg('actions'::text))
  FROM unnest(rs) r
  JOIN install_action_workflow_runs w ON w.id = r.install_action_workflow_run_id
  WHERE w.status = 'finished';
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_runner_jobs() RETURNS trigger AS $$
DECLARE
  rs runner_jobs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::runner_jobs), (n::runner_jobs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.owner_id, o.owner_type, o.type, o.operation, o.runner_id, o.created_at, o.deleted_at)
          IS DISTINCT FROM ROW(n.owner_id, n.owner_type, n.type, n.operation, n.runner_id, n.created_at, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(t.install_id), array_agg(t.partial))
  FROM install_state_runner_job_targets(rs, false) t;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_runner_job_executions() RETURNS trigger AS $$
DECLARE
  rs runner_job_executions[];
  jobs runner_jobs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs
    FROM new_rows n
    WHERE EXISTS (SELECT 1 FROM runner_job_execution_outputs x WHERE x.runner_job_execution_id = n.id);
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::runner_job_executions), (n::runner_job_executions)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.runner_job_id, o.status, o.created_at, o.deleted_at) IS DISTINCT FROM ROW(n.runner_job_id, n.status, n.created_at, n.deleted_at))
      AND EXISTS (SELECT 1 FROM runner_job_execution_outputs x WHERE x.runner_job_execution_id = (v.r).id);
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  SELECT array_agg(j) INTO jobs
  FROM runner_jobs j
  WHERE j.id IN (SELECT r.runner_job_id FROM unnest(rs) r);
  IF jobs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(t.install_id), array_agg(t.partial))
  FROM install_state_runner_job_targets(jobs, true) t;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION install_state_invalidate_runner_job_execution_outputs() RETURNS trigger AS $$
DECLARE
  rs runner_job_execution_outputs[];
  jobs runner_jobs[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::runner_job_execution_outputs), (n::runner_job_execution_outputs)) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL
        OR ROW(o.runner_job_execution_id, o.outputs, o.outputs_blob, o.deleted_at)
          IS DISTINCT FROM ROW(n.runner_job_execution_id, n.outputs, n.outputs_blob, n.deleted_at));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  SELECT array_agg(j) INTO jobs
  FROM runner_jobs j
  WHERE j.id IN (
    SELECT e.runner_job_id
    FROM unnest(rs) r
    JOIN runner_job_executions e ON e.id = r.runner_job_execution_id
  );
  IF jobs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM invalidate_install_state_targets(array_agg(t.install_id), array_agg(t.partial))
  FROM install_state_runner_job_targets(jobs, true) t;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
  source_table text;
BEGIN
  FOREACH source_table IN ARRAY ARRAY[
    'installs',
    'orgs',
    'apps',
    'app_secrets',
    'runner_groups',
    'runners',
    'aws_accounts',
    'azure_accounts',
    'gcp_accounts',
    'app_input_configs',
    'app_inputs',
    'install_inputs',
    'install_sandbox_runs',
    'app_sandbox_configs',
    'public_git_vcs_configs',
    'connected_github_vcs_configs',
    'install_stacks',
    'install_stack_versions',
    'install_stack_outputs',
    'install_components',
    'components',
    'install_deploys',
    'install_action_workflows',
    'action_workflows',
    'install_action_workflow_runs',
    'install_action_workflow_run_steps',
    'runner_jobs',
    'runner_job_executions',
    'runner_job_execution_outputs'
  ] LOOP
    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', 'install_state_invalidate_' || source_table || '_ins', source_table);
    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', 'install_state_invalidate_' || source_table || '_upd', source_table);
    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', 'install_state_invalidate_' || source_table || '_del', source_table);
    EXECUTE format(
      'CREATE TRIGGER %I AFTER INSERT ON %I REFERENCING NEW TABLE AS new_rows FOR EACH STATEMENT EXECUTE FUNCTION %I()',
      'install_state_invalidate_' || source_table || '_ins', source_table, 'install_state_invalidate_' || source_table
    );
    EXECUTE format(
      'CREATE TRIGGER %I AFTER UPDATE ON %I REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows FOR EACH STATEMENT EXECUTE FUNCTION %I()',
      'install_state_invalidate_' || source_table || '_upd', source_table, 'install_state_invalidate_' || source_table
    );
    EXECUTE format(
      'CREATE TRIGGER %I AFTER DELETE ON %I REFERENCING OLD TABLE AS old_rows FOR EACH STATEMENT EXECUTE FUNCTION %I()',
      'install_state_invalidate_' || source_table || '_del', source_table, 'install_state_invalidate_' || source_table
    );
  END LOOP;
END;
$$;

DROP FUNCTION IF EXISTS install_state_runner_job_targets(runner_jobs[], boolean);

CREATE OR REPLACE FUNCTION invalidate_install_state_targets(target_install_ids text[], target_partials text[]) RETURNS void AS $$
DECLARE
  known CONSTANT text[] := {{array .AllPartials}};
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
  JOIN installs i ON i.id = t.install_id
  ORDER BY i.id
  ON CONFLICT (install_id) DO UPDATE SET
    org_id = EXCLUDED.org_id,
    revision = h.revision + 1,
    stale_at = coalesce(h.stale_at, EXCLUDED.stale_at),
    stale_partials = (
      SELECT jsonb_agg(m.partial ORDER BY array_position(known, m.partial), m.partial)
      FROM (
        SELECT jsonb_array_elements_text(h.stale_partials)
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

CREATE OR REPLACE FUNCTION install_state_invalidate_installs() RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    PERFORM invalidate_install_state_targets(array_agg(n.id), array_agg(p.partial))
    FROM new_rows n
    CROSS JOIN unnest({{array .AllPartials}}) AS p(partial);
  ELSE
    PERFORM invalidate_install_state_targets(array_agg(n.id), array_agg(x.partial))
    FROM old_rows o
    JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (
{{- range $i, $c := .InstallColumns}}{{if $i}}
      UNION ALL{{end}}
      SELECT unnest({{array $c.Partials}}) WHERE o.{{$c.Column}} IS DISTINCT FROM n.{{$c.Column}}
{{- end}}
    ) AS x(partial);
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS install_state_invalidate_installs_ins ON installs;
DROP TRIGGER IF EXISTS install_state_invalidate_installs_upd ON installs;
DROP TRIGGER IF EXISTS install_state_invalidate_installs_del ON installs;
CREATE TRIGGER install_state_invalidate_installs_ins AFTER INSERT ON installs REFERENCING NEW TABLE AS new_rows FOR EACH STATEMENT EXECUTE FUNCTION install_state_invalidate_installs();
CREATE TRIGGER install_state_invalidate_installs_upd AFTER UPDATE ON installs REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows FOR EACH STATEMENT EXECUTE FUNCTION install_state_invalidate_installs();
{{range $s := .Sources}}
CREATE OR REPLACE FUNCTION install_state_invalidate_{{.Table}}() RETURNS trigger AS $$
DECLARE
  rs {{.Table}}[];
BEGIN
  IF TG_OP = 'INSERT' THEN
    SELECT array_agg(n) INTO rs FROM new_rows n;
  ELSIF TG_OP = 'DELETE' THEN
    SELECT array_agg(o) INTO rs FROM old_rows o;
  ELSE
    SELECT array_agg(v.r) INTO rs
    FROM old_rows o
    FULL JOIN new_rows n ON n.id = o.id
    CROSS JOIN LATERAL (VALUES (o::{{.Table}}), (n::{{.Table}})) AS v(r)
    WHERE (v.r).id IS NOT NULL
      AND (o.id IS NULL OR n.id IS NULL OR ROW({{columns "o" .Columns}}) IS DISTINCT FROM ROW({{columns "n" .Columns}}));
  END IF;
  IF rs IS NULL THEN
    RETURN NULL;
  END IF;

  PERFORM pg_advisory_xact_lock_shared(hashtext('install_state_sources'), hashtext(o.org_id))
  FROM (
    SELECT DISTINCT {{if eq .Table "orgs"}}r.id{{else}}r.org_id{{end}} AS org_id
    FROM unnest(rs) AS r
  ) o
  ORDER BY o.org_id;

  PERFORM invalidate_install_state_targets(array_agg(t.install_id), array_agg(t.partial))
  FROM (
{{- range $i, $t := .Targets}}{{if $i}}
    UNION ALL{{end}}
    SELECT {{$t.InstallID}}, p.partial
    FROM unnest(rs) AS {{$s.Alias}}{{if $t.Joins}}
    {{$t.Joins}}{{end}}
    CROSS JOIN unnest({{array $t.Partials}}) AS p(partial)
{{- if $t.Where}}
    WHERE {{$t.Where}}{{end}}
{{- end}}
  ) AS t(install_id, partial);
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS install_state_invalidate_{{.Table}}_ins ON {{.Table}};
DROP TRIGGER IF EXISTS install_state_invalidate_{{.Table}}_upd ON {{.Table}};
DROP TRIGGER IF EXISTS install_state_invalidate_{{.Table}}_del ON {{.Table}};
CREATE TRIGGER install_state_invalidate_{{.Table}}_ins AFTER INSERT ON {{.Table}} REFERENCING NEW TABLE AS new_rows FOR EACH STATEMENT EXECUTE FUNCTION install_state_invalidate_{{.Table}}();
CREATE TRIGGER install_state_invalidate_{{.Table}}_upd AFTER UPDATE ON {{.Table}} REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows FOR EACH STATEMENT EXECUTE FUNCTION install_state_invalidate_{{.Table}}();
CREATE TRIGGER install_state_invalidate_{{.Table}}_del AFTER DELETE ON {{.Table}} REFERENCING OLD TABLE AS old_rows FOR EACH STATEMENT EXECUTE FUNCTION install_state_invalidate_{{.Table}}();
{{end}}

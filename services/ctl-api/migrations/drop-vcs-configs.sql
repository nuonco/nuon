ALTER TABLE apps ENABLE ROW LEVEL SECURITY;
CREATE POLICY apps_isolation_policy ON apps FOR select
USING (org_id = current_setting('app.org_id') OR org_id = 'admin');

ALTER TABLE components ENABLE ROW LEVEL SECURITY;
CREATE POLICY components_isolation_policy ON components
USING (org_id = current_setting('app.org_id'));

ALTER TABLE installs ENABLE ROW LEVEL SECURITY;
CREATE POLICY installs_isolation_policy ON installs
USING (org_id = current_setting('app.org_id'));

CREATE OR REPLACE FUNCTION set_org(org_id text) RETURNS void AS $$
BEGIN
    PERFORM set_config('app.org_id', org_id, false);
END;
$$ LANGUAGE plpgsql;

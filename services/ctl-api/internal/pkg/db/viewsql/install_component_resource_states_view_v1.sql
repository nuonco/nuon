-- argMax rather than ROW_NUMBER() OVER: ClickHouse cannot push a WHERE predicate through a
-- window function, so a ranked view full-scans the whole table on every filtered read.
SELECT
    s.org_id AS org_id,
    s.install_id AS install_id,
    s.install_component_id AS install_component_id,
    s.provider AS provider,
    s.api_group AS api_group,
    s.kind AS kind,
    s.namespace AS namespace,
    s.name AS name,
    argMax(s.component_id, s.observed_at) AS component_id,
    argMax(s.runner_id, s.observed_at) AS runner_id,
    argMax(s.source, s.observed_at) AS source,
    argMax(s.owner_name, s.observed_at) AS owner_name,
    argMax(s.health, s.observed_at) AS health,
    argMax(s.message, s.observed_at) AS message,
    argMax(s.native_status, s.observed_at) AS native_status,
    argMax(s.details, s.observed_at) AS details,
    max(s.observed_at) AS observed_at
FROM
    install_component_resource_states AS s
GROUP BY
    s.org_id,
    s.install_id,
    s.install_component_id,
    s.provider,
    s.api_group,
    s.kind,
    s.namespace,
    s.name;

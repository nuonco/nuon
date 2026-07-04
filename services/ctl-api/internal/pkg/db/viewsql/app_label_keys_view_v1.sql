SELECT
  key,
  values,
  entity_types,
  usage_count,
  first_used_at,
  (ROW_NUMBER() OVER (PARTITION BY app_id ORDER BY first_used_at ASC) - 1)::int AS color_index,
  app_id,
  org_id
FROM (
  SELECT
    key,
    array_agg(DISTINCT value ORDER BY value)             AS values,
    array_agg(DISTINCT entity_type ORDER BY entity_type) AS entity_types,
    count(*)::int                                        AS usage_count,
    min(first_used_at)                                   AS first_used_at,
    app_id,
    org_id
  FROM (
    SELECT
      (jsonb_each_text(labels)).key   AS key,
      (jsonb_each_text(labels)).value AS value,
      'component'                     AS entity_type,
      app_id,
      (SELECT org_id FROM apps WHERE apps.id = components.app_id AND apps.deleted_at = 0 LIMIT 1) AS org_id,
      created_at                      AS first_used_at
    FROM components
    WHERE labels IS NOT NULL
      AND labels != '{}'::jsonb
      AND deleted_at = 0

    UNION ALL

    SELECT
      (jsonb_each_text(labels)).key   AS key,
      (jsonb_each_text(labels)).value AS value,
      'action'                        AS entity_type,
      app_id,
      org_id,
      created_at                      AS first_used_at
    FROM action_workflows
    WHERE labels IS NOT NULL
      AND labels != '{}'::jsonb
      AND deleted_at = 0

    UNION ALL

    SELECT
      (jsonb_each_text(labels)).key   AS key,
      (jsonb_each_text(labels)).value AS value,
      'runbook'                       AS entity_type,
      app_id,
      org_id,
      created_at                      AS first_used_at
    FROM runbooks
    WHERE labels IS NOT NULL
      AND labels != '{}'::jsonb
      AND deleted_at = 0

    UNION ALL

    SELECT
      (jsonb_each_text(labels)).key   AS key,
      (jsonb_each_text(labels)).value AS value,
      'install'                       AS entity_type,
      app_id,
      (SELECT org_id FROM apps WHERE apps.id = installs.app_id AND apps.deleted_at = 0 LIMIT 1) AS org_id,
      created_at                      AS first_used_at
    FROM installs
    WHERE labels IS NOT NULL
      AND labels != '{}'::jsonb
      AND deleted_at = 0
  ) AS label_entries
  GROUP BY key, app_id, org_id
) AS grouped_labels
ORDER BY first_used_at ASC

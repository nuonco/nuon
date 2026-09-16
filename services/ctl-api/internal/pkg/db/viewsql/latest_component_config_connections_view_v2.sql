SELECT l.*,
  (SELECT count(*) FROM component_config_connections c2
    WHERE c2.component_id = l.component_id AND (c2.created_at, c2.id) <= (l.created_at, l.id)) AS version
FROM (SELECT DISTINCT ON (ccc.component_id) ccc.*
      FROM component_config_connections ccc
      ORDER BY ccc.component_id, ccc.created_at DESC, ccc.id DESC) l

DROP VIEW IF EXISTS table_sizes_view_v1;

CREATE
OR REPLACE VIEW table_sizes_view_v1 AS
  SELECT
    table_name,
    pg_size_pretty(pg_total_relation_size(quote_ident(table_name))) as size_human,
    pg_total_relation_size(quote_ident(table_name)) as size_bytes
  FROM information_schema.tables
  WHERE table_schema = 'public'
  ORDER BY 3 desc;

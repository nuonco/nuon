-- Fix incorrectly defined Map skip indexes on otel_log_records.
--
-- The original GORM annotations had two classes of bugs:
--   1) idx_res_attr_value used mapKeys() instead of mapValues()
--   2) idx_scope_attr_{key,value} and idx_log_attr_{key,value} all referenced
--      mapKeys(resource_attributes) instead of their own column.
--
-- Drop the broken indexes and recreate them pointing at the correct
-- column / function. Then MATERIALIZE on existing parts so historical
-- data also benefits (TTL still drops anything older than 30 days).

-- resource_attributes value index (was mapKeys, should be mapValues)
ALTER TABLE otel_log_records ON CLUSTER simple DROP INDEX IF EXISTS idx_res_attr_value;
ALTER TABLE otel_log_records ON CLUSTER simple
    ADD INDEX idx_res_attr_value mapValues(resource_attributes) TYPE bloom_filter(0.1) GRANULARITY 1;
ALTER TABLE otel_log_records ON CLUSTER simple MATERIALIZE INDEX idx_res_attr_value;

-- scope_attributes key + value indexes (were both pointing at resource_attributes)
ALTER TABLE otel_log_records ON CLUSTER simple DROP INDEX IF EXISTS idx_scope_attr_key;
ALTER TABLE otel_log_records ON CLUSTER simple
    ADD INDEX idx_scope_attr_key mapKeys(scope_attributes) TYPE bloom_filter(0.1) GRANULARITY 1;
ALTER TABLE otel_log_records ON CLUSTER simple MATERIALIZE INDEX idx_scope_attr_key;

ALTER TABLE otel_log_records ON CLUSTER simple DROP INDEX IF EXISTS idx_scope_attr_value;
ALTER TABLE otel_log_records ON CLUSTER simple
    ADD INDEX idx_scope_attr_value mapValues(scope_attributes) TYPE bloom_filter(0.1) GRANULARITY 1;
ALTER TABLE otel_log_records ON CLUSTER simple MATERIALIZE INDEX idx_scope_attr_value;

-- log_attributes key + value indexes (were both pointing at resource_attributes)
ALTER TABLE otel_log_records ON CLUSTER simple DROP INDEX IF EXISTS idx_log_attr_key;
ALTER TABLE otel_log_records ON CLUSTER simple
    ADD INDEX idx_log_attr_key mapKeys(log_attributes) TYPE bloom_filter(0.1) GRANULARITY 1;
ALTER TABLE otel_log_records ON CLUSTER simple MATERIALIZE INDEX idx_log_attr_key;

ALTER TABLE otel_log_records ON CLUSTER simple DROP INDEX IF EXISTS idx_log_attr_value;
ALTER TABLE otel_log_records ON CLUSTER simple
    ADD INDEX idx_log_attr_value mapValues(log_attributes) TYPE bloom_filter(0.1) GRANULARITY 1;
ALTER TABLE otel_log_records ON CLUSTER simple MATERIALIZE INDEX idx_log_attr_value;

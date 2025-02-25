SELECT 
    table AS table_name,
    formatReadableSize(sum(bytes)) AS size_pretty,
    sum(bytes) AS size_bytes
FROM system.parts
WHERE active AND database = 'ctl_api'
GROUP BY table
ORDER BY size_bytes DESC;


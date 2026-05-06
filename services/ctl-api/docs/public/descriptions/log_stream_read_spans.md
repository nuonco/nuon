Read OTEL trace spans for a log stream.

Returns the flat list of spans recorded by the runner for the job execution
(or executions) associated with this log stream, ordered by start timestamp
ASC. The frontend assembles the tree from `parent_span_id`.

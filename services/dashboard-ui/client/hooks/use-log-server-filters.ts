import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import type { TLogStreamFilters } from '@/lib/ctl-api/log-streams/get-log-stream-logs'

// Reads the same URL params as useLogFilters and returns them in the shape
// expected by the ctl-api log-stream read endpoint. Kept as a standalone
// hook so the unified logs provider (which does the fetching) can read the
// filter state without depending on the consumer-facing useLogFilters hook.
export const useLogServerFilters = (): TLogStreamFilters => {
  const [searchParams] = useSearchParams()

  return useMemo(() => {
    const f: TLogStreamFilters = {}

    const includeSystemLogs = searchParams.get('system_logs') !== 'false'
    if (!includeSystemLogs) f.scope_name = ['oteljob']

    const severities = searchParams.getAll('severity')
    if (severities.length > 0) f.severity_text = severities

    const tool = searchParams.get('tool')
    if (tool) f.tool = tool

    const helmReleaseName = searchParams.get('helm_release_name')
    if (helmReleaseName) f.helm_release_name = helmReleaseName

    const helmOperation = searchParams.get('helm_operation')
    if (helmOperation) f.helm_operation = helmOperation

    const tfWorkspaceID = searchParams.get('tf_workspace_id')
    if (tfWorkspaceID) f.tf_workspace_id = tfWorkspaceID

    const tfOperation = searchParams.get('tf_operation')
    if (tfOperation) f.tf_operation = tfOperation

    const k8sKind = searchParams.get('k8s_kind')
    if (k8sKind) f.k8s_kind = k8sKind

    const k8sNamespace = searchParams.get('k8s_namespace')
    if (k8sNamespace) f.k8s_namespace = k8sNamespace

    const k8sName = searchParams.get('k8s_name')
    if (k8sName) f.k8s_name = k8sName

    const q = searchParams.get('q')
    if (q && q.trim()) f.q = q.trim()

    const spanId = searchParams.get('span_id')
    if (spanId) f.span_id = spanId

    const traceId = searchParams.get('trace_id')
    if (traceId) f.trace_id = traceId

    return f
  }, [searchParams])
}

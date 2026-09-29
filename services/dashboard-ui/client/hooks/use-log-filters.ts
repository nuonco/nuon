import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router'
import type { TLogStreamFilters } from '@/lib/ctl-api/log-streams/get-log-stream-logs'
import type { TOTELLog, TSpan } from '@/types'
import { collectDescendantIds } from '@/utils/span-tree'

type SortDirection = 'asc' | 'desc'
export type ViewMode = 'structured' | 'raw'

const PARAM_SEVERITY = 'severity'
const PARAM_TOOL = 'tool'
const PARAM_HELM_RELEASE = 'helm_release_name'
const PARAM_HELM_OPERATION = 'helm_operation'
const PARAM_TF_WORKSPACE = 'tf_workspace_id'
const PARAM_TF_OPERATION = 'tf_operation'
const PARAM_K8S_KIND = 'k8s_kind'
const PARAM_K8S_NAMESPACE = 'k8s_namespace'
const PARAM_K8S_NAME = 'k8s_name'
const PARAM_BODY = 'q'
const PARAM_SYSTEM_LOGS = 'system_logs'
const PARAM_SORT = 'sort'
const PARAM_VIEW = 'view'
// why: Span / trace cross-link from the trace tab. trace_id is an exact server-side
// match; span_id stays client-side because the UI expands a parent span into
// its descendants before filtering, which the server can't do.
const PARAM_SPAN_ID = 'span_id'
const PARAM_TRACE_ID = 'trace_id'

// why: URL filter state mapped onto the ctl-api log read/tail query params, so
// filtering happens server-side instead of only in the browser. span_id is
// deliberately excluded — the server can only exact-match it, while the UI
// expands a parent span into its descendants before filtering.
export const buildServerFilters = (
  searchParams: URLSearchParams
): TLogStreamFilters => {
  const filters: TLogStreamFilters = {}

  const severities = searchParams.getAll(PARAM_SEVERITY)
  if (severities.length === 0 && DEFAULT_SEVERITIES.length > 0) {
    filters.severity_text = [...DEFAULT_SEVERITIES]
  } else if (severities.length > 0) {
    filters.severity_text = severities
  }

  if (searchParams.get(PARAM_SYSTEM_LOGS) === 'false') {
    filters.scope_name = ['oteljob']
  }

  const single = (urlKey: string, apiKey: keyof TLogStreamFilters) => {
    const v = searchParams.get(urlKey)
    if (v) (filters[apiKey] as string) = v
  }
  single(PARAM_TOOL, 'tool')
  single(PARAM_HELM_RELEASE, 'helm_release_name')
  single(PARAM_HELM_OPERATION, 'helm_operation')
  single(PARAM_TF_WORKSPACE, 'tf_workspace_id')
  single(PARAM_TF_OPERATION, 'tf_operation')
  single(PARAM_K8S_KIND, 'k8s_kind')
  single(PARAM_K8S_NAMESPACE, 'k8s_namespace')
  single(PARAM_K8S_NAME, 'k8s_name')
  single(PARAM_TRACE_ID, 'trace_id')

  const q = searchParams.get(PARAM_BODY)
  if (q && q.trim()) filters.q = q.trim()

  return filters
}

const ALL_FILTER_PARAMS = [
  PARAM_SEVERITY,
  PARAM_TOOL,
  PARAM_HELM_RELEASE,
  PARAM_HELM_OPERATION,
  PARAM_TF_WORKSPACE,
  PARAM_TF_OPERATION,
  PARAM_K8S_KIND,
  PARAM_K8S_NAMESPACE,
  PARAM_K8S_NAME,
  PARAM_BODY,
  PARAM_SYSTEM_LOGS,
] as const

const KNOWN_SEVERITIES = ['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal']

const DEFAULT_SEVERITIES = ['Info', 'Warn', 'Error', 'Fatal']

export const useLogFilters = <T extends TOTELLog>(
  logs: T[] | null,
  spans?: TSpan[],
  streamId?: string
) => {
  const [searchParams, setSearchParams] = useSearchParams()

  const selectedSeverities = useMemo(() => {
    const fromURL = searchParams.getAll(PARAM_SEVERITY)
    if (fromURL.length === 0) return new Set(DEFAULT_SEVERITIES)
    return new Set(fromURL)
  }, [searchParams])
  const severityIsDefault = searchParams.getAll(PARAM_SEVERITY).length === 0
  const includeSystemLogs = searchParams.get(PARAM_SYSTEM_LOGS) !== 'false'
  const tool = searchParams.get(PARAM_TOOL) || ''
  const helmReleaseName = searchParams.get(PARAM_HELM_RELEASE) || ''
  const helmOperation = searchParams.get(PARAM_HELM_OPERATION) || ''
  const tfWorkspaceID = searchParams.get(PARAM_TF_WORKSPACE) || ''
  const tfOperation = searchParams.get(PARAM_TF_OPERATION) || ''
  const k8sKind = searchParams.get(PARAM_K8S_KIND) || ''
  const k8sNamespace = searchParams.get(PARAM_K8S_NAMESPACE) || ''
  const k8sName = searchParams.get(PARAM_K8S_NAME) || ''
  const searchQuery = searchParams.get(PARAM_BODY) || ''
  const spanId = searchParams.get(PARAM_SPAN_ID) || ''
  const traceId = searchParams.get(PARAM_TRACE_ID) || ''
  const sortDirection: SortDirection =
    searchParams.get(PARAM_SORT) === 'asc' ? 'asc' : 'desc'
  const viewMode: ViewMode =
    searchParams.get(PARAM_VIEW) === 'raw' ? 'raw' : 'structured'

  const updateParams = useCallback(
    (mutate: (sp: URLSearchParams) => void) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          mutate(next)
          return next
        },
        { replace: true }
      )
    },
    [setSearchParams]
  )

  const setMultiValue = useCallback(
    (key: string, values: Set<string> | string[]) => {
      updateParams((next) => {
        next.delete(key)
        for (const v of values) {
          if (v) next.append(key, v)
        }
      })
    },
    [updateParams]
  )

  const setSingleValue = useCallback(
    (key: string, value: string) => {
      updateParams((next) => {
        if (!value) next.delete(key)
        else next.set(key, value)
      })
    },
    [updateParams]
  )

  // why: Available values are derived from the logs loaded so far for this
  // stream — not just the current page. Once `tool` filters server-side,
  // newly loaded rows only ever contain the selected tool, so deriving
  // the facet from the live set alone would collapse the dropdown to the
  // selection and force the user to clear the filter to switch.
  //
  // The accumulation resets on stream identity, NOT on logs being empty:
  // applying a server-side filter reconnects the stream and clears logs,
  // which must not wipe facets seen earlier on the same stream. When the
  // stream id changes, the logs in that first render may still belong to
  // the old stream, so the new stream starts with an empty set and
  // accumulates from its own rows only.
  const [toolFacets, setToolFacets] = useState<{
    streamId?: string
    tools: Set<string>
  }>({ tools: new Set() })
  useEffect(() => {
    setToolFacets((prev) => {
      if (
        streamId !== undefined &&
        prev.streamId !== undefined &&
        prev.streamId !== streamId
      ) {
        // why: Stream changed: this render's logs may still belong to the old
        // stream, so start empty without accumulating. Recording the new id
        // here (even when logs are empty) matters — the render that carries
        // the id change can only hold old-stream logs or [], so dropping
        // them is safe, and the reset can never fire on real new-stream
        // data.
        return { streamId, tools: new Set<string>() }
      }
      let changed = prev.streamId !== streamId
      const tools = new Set(prev.tools)
      for (const log of logs ?? []) {
        const t = log.log_attributes?.['nuon.tool']
        if (t && !tools.has(t)) {
          tools.add(t)
          changed = true
        }
      }
      return changed ? { streamId, tools } : prev
    })
  }, [logs, streamId])

  const availableTools = toolFacets.tools

  const availableSeverities = useMemo(() => {
    const out = new Set<string>(KNOWN_SEVERITIES)
    if (logs) {
      for (const log of logs) {
        if (log.severity_text) out.add(log.severity_text)
      }
    }
    return out
  }, [logs])

  const sortLogsByTimestamp = (records: T[], direction: SortDirection): T[] => {
    return [...records].sort((a, b) => {
      const aTimestamp = a.timestamp
      const bTimestamp = b.timestamp
      if (direction === 'desc') {
        return bTimestamp > aTimestamp ? 1 : bTimestamp < aTimestamp ? -1 : 0
      }
      return aTimestamp > bTimestamp ? 1 : aTimestamp < bTimestamp ? -1 : 0
    })
  }

  const spanIdMatchSet = useMemo(() => {
    if (!spanId) return new Set<string>()
    if (spans && spans.length > 0) return collectDescendantIds(spans, spanId)
    return new Set<string>([spanId])
  }, [spanId, spans])

  const filteredLogs = useMemo(() => {
    if (!logs) return null

    let filtered = logs

    if (selectedSeverities.size > 0) {
      filtered = filtered.filter((item) =>
        selectedSeverities.has(item.severity_text)
      )
    }

    if (!includeSystemLogs) {
      filtered = filtered.filter((item) => item.scope_name === 'oteljob')
    }

    if (tool) {
      filtered = filtered.filter((item) => item.log_attributes?.['nuon.tool'] === tool)
    }
    if (helmReleaseName) {
      filtered = filtered.filter((item) => item.log_attributes?.['helm.release_name'] === helmReleaseName)
    }
    if (helmOperation) {
      filtered = filtered.filter((item) => item.log_attributes?.['helm.operation'] === helmOperation)
    }
    if (tfWorkspaceID) {
      filtered = filtered.filter((item) => item.log_attributes?.['tf.workspace_id'] === tfWorkspaceID)
    }
    if (tfOperation) {
      filtered = filtered.filter((item) => item.log_attributes?.['tf.operation'] === tfOperation)
    }
    if (k8sKind) {
      filtered = filtered.filter((item) => item.log_attributes?.['k8s.kind'] === k8sKind)
    }
    if (k8sNamespace) {
      filtered = filtered.filter((item) => item.log_attributes?.['k8s.namespace'] === k8sNamespace)
    }
    if (k8sName) {
      filtered = filtered.filter((item) => item.log_attributes?.['k8s.name'] === k8sName)
    }

    if (spanId) {
      filtered = filtered.filter((item) => spanIdMatchSet.has(item.span_id))
    }
    if (traceId) {
      filtered = filtered.filter((item) => item.trace_id === traceId)
    }

    if (searchQuery.trim()) {
      const searchLower = searchQuery.toLowerCase().trim()
      filtered = filtered.filter((item) =>
        item.body?.toLowerCase().includes(searchLower)
      )
    }

    return sortLogsByTimestamp(filtered, sortDirection)
  }, [
    logs,
    selectedSeverities,
    includeSystemLogs,
    tool,
    helmReleaseName,
    helmOperation,
    tfWorkspaceID,
    tfOperation,
    k8sKind,
    k8sNamespace,
    k8sName,
    spanIdMatchSet,
    traceId,
    searchQuery,
    sortDirection,
  ])

  const handleSeverityInputToggle = useCallback(
    (severity: string) => {
      const next = new Set(selectedSeverities)
      if (next.has(severity)) next.delete(severity)
      else next.add(severity)
      setMultiValue(PARAM_SEVERITY, next)
    },
    [selectedSeverities, setMultiValue]
  )
  const handleSeverityButtonClick = useCallback(
    (severity: string) => {
      if (selectedSeverities.size === 1 && selectedSeverities.has(severity)) {
        setMultiValue(PARAM_SEVERITY, [])
      } else {
        setMultiValue(PARAM_SEVERITY, [severity])
      }
    },
    [selectedSeverities, setMultiValue]
  )
  const handleSeverityReset = useCallback(() => {
    setMultiValue(PARAM_SEVERITY, [])
  }, [setMultiValue])

  const setTool = useCallback((v: string) => setSingleValue(PARAM_TOOL, v), [setSingleValue])
  const setHelmReleaseName = useCallback((v: string) => setSingleValue(PARAM_HELM_RELEASE, v), [setSingleValue])
  const setHelmOperation = useCallback((v: string) => setSingleValue(PARAM_HELM_OPERATION, v), [setSingleValue])
  const setTfWorkspaceID = useCallback((v: string) => setSingleValue(PARAM_TF_WORKSPACE, v), [setSingleValue])
  const setTfOperation = useCallback((v: string) => setSingleValue(PARAM_TF_OPERATION, v), [setSingleValue])
  const setK8sKind = useCallback((v: string) => setSingleValue(PARAM_K8S_KIND, v), [setSingleValue])
  const setK8sNamespace = useCallback((v: string) => setSingleValue(PARAM_K8S_NAMESPACE, v), [setSingleValue])
  const setK8sName = useCallback((v: string) => setSingleValue(PARAM_K8S_NAME, v), [setSingleValue])

  const handleSearchChange = useCallback(
    (query: string) => setSingleValue(PARAM_BODY, query),
    [setSingleValue]
  )
  const handleSortToggle = useCallback(() => {
    setSingleValue(PARAM_SORT, sortDirection === 'desc' ? 'asc' : 'desc')
  }, [sortDirection, setSingleValue])
  const handleSortChange = useCallback(
    (direction: SortDirection) => setSingleValue(PARAM_SORT, direction),
    [setSingleValue]
  )
  const handleViewModeChange = useCallback(
    (mode: ViewMode) => {
      updateParams((next) => {
        if (mode === 'raw') {
          next.set(PARAM_VIEW, 'raw')
          next.set(PARAM_SORT, 'asc')
          next.set(PARAM_SYSTEM_LOGS, 'false')
        } else {
          next.delete(PARAM_VIEW)
          next.delete(PARAM_SORT)
          next.delete(PARAM_SYSTEM_LOGS)
        }
      })
    },
    [updateParams]
  )

  const handleSystemLogsToggle = useCallback(() => {
    updateParams((next) => {
      if (includeSystemLogs) next.set(PARAM_SYSTEM_LOGS, 'false')
      else next.delete(PARAM_SYSTEM_LOGS)
    })
  }, [includeSystemLogs, updateParams])

  const isFiltered =
    !severityIsDefault ||
    !includeSystemLogs ||
    !!tool ||
    !!helmReleaseName ||
    !!helmOperation ||
    !!tfWorkspaceID ||
    !!tfOperation ||
    !!k8sKind ||
    !!k8sNamespace ||
    !!k8sName ||
    searchQuery.trim() !== ''

  const handleResetAll = useCallback(() => {
    updateParams((next) => {
      for (const key of ALL_FILTER_PARAMS) next.delete(key)
    })
  }, [updateParams])

  const serverFilters: TLogStreamFilters = useMemo(
    () => buildServerFilters(searchParams),
    [searchParams]
  )

  return {
    selectedSeverities,
    availableSeverities,
    handleSeverityInputToggle,
    handleSeverityButtonClick,
    handleSeverityReset,

    includeSystemLogs,
    handleSystemLogsToggle,

    availableTools,
    tool,
    setTool,
    helmReleaseName,
    setHelmReleaseName,
    helmOperation,
    setHelmOperation,
    tfWorkspaceID,
    setTfWorkspaceID,
    tfOperation,
    setTfOperation,
    k8sKind,
    setK8sKind,
    k8sNamespace,
    setK8sNamespace,
    k8sName,
    setK8sName,

    viewMode,
    handleViewModeChange,

    searchQuery,
    sortDirection,
    filteredLogs,
    handleSearchChange,
    handleSortToggle,
    handleSortChange,

    isFiltered,
    handleResetAll,

    serverFilters,

    filterStats: {
      selectedCount: filteredLogs?.length || 0,
      totalCount: logs?.length || 0,
    },
    sortStats: {
      direction: sortDirection,
      isNewestFirst: sortDirection === 'desc',
      isOldestFirst: sortDirection === 'asc',
    },
    severityStats: {
      selectedCount: selectedSeverities.size,
      totalCount: availableSeverities.size,
      isDefault: severityIsDefault,
    },
  }
}

export type TLogFiltersProps = ReturnType<typeof useLogFilters>

import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router'
import type { TLogStreamFilters } from '@/lib/ctl-api/log-streams/get-log-stream-logs'
import type { TOTELLog, TSpan } from '@/types'
import { collectDescendantIds } from '@/utils/span-tree'

type SortDirection = 'asc' | 'desc'

// URL-backed filter state. Empty selection (no entries in the URL) means
// "show everything" — we never pre-filter.
//
// Multi-value params (severity, service, scope) are repeated in the URL,
// e.g. ?severity=Info&severity=Warn. Single-value attribute filters
// (tool, helm.release_name, k8s.kind, etc.) appear once.
const PARAM_SEVERITY = 'severity'
const PARAM_SERVICE = 'service'
const PARAM_SCOPE = 'scope'
const PARAM_TOOL = 'tool'
const PARAM_HELM_RELEASE = 'helm_release_name'
const PARAM_HELM_OPERATION = 'helm_operation'
const PARAM_TF_WORKSPACE = 'tf_workspace_id'
const PARAM_TF_OPERATION = 'tf_operation'
const PARAM_K8S_KIND = 'k8s_kind'
const PARAM_K8S_NAMESPACE = 'k8s_namespace'
const PARAM_K8S_NAME = 'k8s_name'
const PARAM_BODY = 'q'
const PARAM_JOB_OUTPUT = 'job_output'
const PARAM_SORT = 'sort'
// Span / trace cross-link from the trace tab. SSE delivers every log on the
// stream regardless of server filter, so we always have to apply these
// client-side for the cross-link to feel snappy. The static-fetch path
// (useLogServerFilters) also forwards these to ctl-api.
const PARAM_SPAN_ID = 'span_id'
const PARAM_TRACE_ID = 'trace_id'

const ALL_FILTER_PARAMS = [
  PARAM_SEVERITY,
  PARAM_SERVICE,
  PARAM_SCOPE,
  PARAM_TOOL,
  PARAM_HELM_RELEASE,
  PARAM_HELM_OPERATION,
  PARAM_TF_WORKSPACE,
  PARAM_TF_OPERATION,
  PARAM_K8S_KIND,
  PARAM_K8S_NAMESPACE,
  PARAM_K8S_NAME,
  PARAM_BODY,
  PARAM_JOB_OUTPUT,
] as const

const KNOWN_SEVERITIES = ['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal']

// Severity has a sensible default — Trace/Debug are very noisy and almost
// always want hiding out of the box. When no `severity` param is present in
// the URL we apply this set; once the user toggles anything we honor exactly
// what they asked for. handleSeverityReset clears the URL param to return
// to defaults.
const DEFAULT_SEVERITIES = ['Info', 'Warn', 'Error', 'Fatal']

export const useLogFilters = <T extends TOTELLog>(
  logs: T[] | null,
  // Optional span list for parent-aggregation in the span→logs cross-link.
  // When provided, a single ?span_id=X URL param is expanded into the set
  // { X, ...descendants(X) } before filtering, so clicking a parent step
  // span (e.g. step.execute) shows logs from all child tool spans too.
  // When omitted (most callers), span_id behaves as an exact-match filter
  // — preserving the original behavior for surfaces that don't render the
  // trace tab.
  spans?: TSpan[]
) => {
  const [searchParams, setSearchParams] = useSearchParams()

  const selectedSeverities = useMemo(() => {
    const fromURL = searchParams.getAll(PARAM_SEVERITY)
    if (fromURL.length === 0) return new Set(DEFAULT_SEVERITIES)
    return new Set(fromURL)
  }, [searchParams])
  const severityIsDefault = searchParams.getAll(PARAM_SEVERITY).length === 0
  const selectedServices = useMemo(
    () => new Set(searchParams.getAll(PARAM_SERVICE)),
    [searchParams]
  )
  const selectedScopes = useMemo(
    () => new Set(searchParams.getAll(PARAM_SCOPE)),
    [searchParams]
  )
  const tool = searchParams.get(PARAM_TOOL) || ''
  const helmReleaseName = searchParams.get(PARAM_HELM_RELEASE) || ''
  const helmOperation = searchParams.get(PARAM_HELM_OPERATION) || ''
  const tfWorkspaceID = searchParams.get(PARAM_TF_WORKSPACE) || ''
  const tfOperation = searchParams.get(PARAM_TF_OPERATION) || ''
  const k8sKind = searchParams.get(PARAM_K8S_KIND) || ''
  const k8sNamespace = searchParams.get(PARAM_K8S_NAMESPACE) || ''
  const k8sName = searchParams.get(PARAM_K8S_NAME) || ''
  const searchQuery = searchParams.get(PARAM_BODY) || ''
  const jobOutputOnly = searchParams.get(PARAM_JOB_OUTPUT) === 'true'
  const spanId = searchParams.get(PARAM_SPAN_ID) || ''
  const traceId = searchParams.get(PARAM_TRACE_ID) || ''
  const sortDirection: SortDirection =
    searchParams.get(PARAM_SORT) === 'asc' ? 'asc' : 'desc'

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

  // Available values are derived from the currently loaded logs; useful for
  // populating dropdowns that haven't been pre-seeded with a fixed list.
  const availableServices = useMemo(() => {
    const out = new Set<string>()
    if (!logs) return out
    for (const log of logs) {
      if (log.service_name) out.add(log.service_name)
    }
    return out
  }, [logs])

  const availableScopes = useMemo(() => {
    const out = new Set<string>()
    if (!logs) return out
    for (const log of logs) {
      if (log.scope_name) out.add(log.scope_name)
    }
    return out
  }, [logs])

  const availableTools = useMemo(() => {
    const out = new Set<string>()
    if (!logs) return out
    for (const log of logs) {
      const t = log.log_attributes?.['nuon.tool']
      if (t) out.add(t)
    }
    return out
  }, [logs])

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

  // Span-id match set for the cross-link. When a span list is provided we
  // expand the URL's span_id into the full descendant set; otherwise we fall
  // back to exact match. The filter callback below reads `spanIdMatchSet`.
  const spanIdMatchSet = useMemo(() => {
    if (!spanId) return new Set<string>()
    if (spans && spans.length > 0) return collectDescendantIds(spans, spanId)
    return new Set<string>([spanId])
  }, [spanId, spans])

  const filteredLogs = useMemo(() => {
    if (!logs) return null

    let filtered = logs

    // Empty selection == show everything for these multi-value filters.
    if (selectedSeverities.size > 0) {
      filtered = filtered.filter((item) =>
        selectedSeverities.has(item.severity_text)
      )
    }
    if (selectedServices.size > 0) {
      filtered = filtered.filter((item) =>
        item.service_name ? selectedServices.has(item.service_name) : false
      )
    }
    if (selectedScopes.size > 0) {
      filtered = filtered.filter((item) =>
        item.scope_name ? selectedScopes.has(item.scope_name) : false
      )
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

    if (jobOutputOnly) {
      filtered = filtered.filter((item) => item.scope_name === 'oteljob')
    }

    if (spanId) {
      // Expand spanId into { spanId, ...descendants } when a span list is
      // available so clicking a parent step span shows all child tool span
      // logs too. Without spans we fall back to exact match.
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
    selectedServices,
    selectedScopes,
    tool,
    helmReleaseName,
    helmOperation,
    tfWorkspaceID,
    tfOperation,
    k8sKind,
    k8sNamespace,
    k8sName,
    jobOutputOnly,
    spanIdMatchSet,
    traceId,
    searchQuery,
    sortDirection,
  ])

  // Severity handlers
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
      // "Only" semantics: pin selection to just this one. If already
      // pinned to it, clear (== show all).
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

  // Service handlers
  const handleServiceInputToggle = useCallback(
    (service: string) => {
      const next = new Set(selectedServices)
      if (next.has(service)) next.delete(service)
      else next.add(service)
      setMultiValue(PARAM_SERVICE, next)
    },
    [selectedServices, setMultiValue]
  )
  const handleServiceButtonClick = useCallback(
    (service: string) => {
      if (selectedServices.size === 1 && selectedServices.has(service)) {
        setMultiValue(PARAM_SERVICE, [])
      } else {
        setMultiValue(PARAM_SERVICE, [service])
      }
    },
    [selectedServices, setMultiValue]
  )
  const handleServiceReset = useCallback(() => {
    setMultiValue(PARAM_SERVICE, [])
  }, [setMultiValue])

  // Scope handlers
  const handleScopeInputToggle = useCallback(
    (scope: string) => {
      const next = new Set(selectedScopes)
      if (next.has(scope)) next.delete(scope)
      else next.add(scope)
      setMultiValue(PARAM_SCOPE, next)
    },
    [selectedScopes, setMultiValue]
  )
  const handleScopeButtonClick = useCallback(
    (scope: string) => {
      if (selectedScopes.size === 1 && selectedScopes.has(scope)) {
        setMultiValue(PARAM_SCOPE, [])
      } else {
        setMultiValue(PARAM_SCOPE, [scope])
      }
    },
    [selectedScopes, setMultiValue]
  )
  const handleScopeReset = useCallback(() => {
    setMultiValue(PARAM_SCOPE, [])
  }, [setMultiValue])

  // Single-value attribute setters
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
  const handleJobOutputToggle = useCallback(() => {
    updateParams((next) => {
      if (jobOutputOnly) next.delete(PARAM_JOB_OUTPUT)
      else next.set(PARAM_JOB_OUTPUT, 'true')
    })
  }, [jobOutputOnly, updateParams])

  const isFiltered =
    !severityIsDefault ||
    selectedServices.size > 0 ||
    selectedScopes.size > 0 ||
    !!tool ||
    !!helmReleaseName ||
    !!helmOperation ||
    !!tfWorkspaceID ||
    !!tfOperation ||
    !!k8sKind ||
    !!k8sNamespace ||
    !!k8sName ||
    jobOutputOnly ||
    searchQuery.trim() !== ''

  const handleResetAll = useCallback(() => {
    updateParams((next) => {
      for (const key of ALL_FILTER_PARAMS) next.delete(key)
    })
  }, [updateParams])

  // serverFilters mirrors the URL state in the shape the ctl-api endpoint
  // expects, so callers can pass it straight to getLogStreamLogs[WithMeta].
  const serverFilters: TLogStreamFilters = useMemo(() => {
    const f: TLogStreamFilters = {}
    if (selectedServices.size > 0) f.service_name = Array.from(selectedServices)
    if (selectedScopes.size > 0) f.scope_name = Array.from(selectedScopes)
    if (selectedSeverities.size > 0) f.severity_text = Array.from(selectedSeverities)
    if (tool) f.tool = tool
    if (helmReleaseName) f.helm_release_name = helmReleaseName
    if (helmOperation) f.helm_operation = helmOperation
    if (tfWorkspaceID) f.tf_workspace_id = tfWorkspaceID
    if (tfOperation) f.tf_operation = tfOperation
    if (k8sKind) f.k8s_kind = k8sKind
    if (k8sNamespace) f.k8s_namespace = k8sNamespace
    if (k8sName) f.k8s_name = k8sName
    if (searchQuery.trim()) f.q = searchQuery.trim()
    return f
  }, [
    selectedServices,
    selectedScopes,
    selectedSeverities,
    tool,
    helmReleaseName,
    helmOperation,
    tfWorkspaceID,
    tfOperation,
    k8sKind,
    k8sNamespace,
    k8sName,
    searchQuery,
  ])

  return {
    // Severity filter
    selectedSeverities,
    availableSeverities,
    handleSeverityInputToggle,
    handleSeverityButtonClick,
    handleSeverityReset,

    // Service filter
    selectedServices,
    availableServices,
    handleServiceInputToggle,
    handleServiceButtonClick,
    handleServiceReset,

    // Scope filter
    selectedScopes,
    availableScopes,
    handleScopeInputToggle,
    handleScopeButtonClick,
    handleScopeReset,

    // Tool / attribute filters
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

    // Job output filter
    jobOutputOnly,
    handleJobOutputToggle,

    // Search and sort
    searchQuery,
    sortDirection,
    filteredLogs,
    handleSearchChange,
    handleSortToggle,
    handleSortChange,

    // Reset
    isFiltered,
    handleResetAll,

    // Server-side filter shape (pass to getLogStreamLogs)
    serverFilters,

    // Stats
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
      // "Is the user currently on the default selection (no URL override)?"
      isDefault: severityIsDefault,
    },
    serviceStats: {
      selectedCount: selectedServices.size,
      totalCount: availableServices.size,
      isAllSelected: selectedServices.size === 0,
    },
    scopeStats: {
      selectedCount: selectedScopes.size,
      totalCount: availableScopes.size,
      isAllSelected: selectedScopes.size === 0,
    },
  }
}

export type TLogFiltersProps = ReturnType<typeof useLogFilters>

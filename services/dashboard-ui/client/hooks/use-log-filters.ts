import { useEffect, useMemo, useState } from 'react'
import type { TOTELLog } from '@/types'

type SortDirection = 'asc' | 'desc'

const DEFAULT_SELECTED_SEVERITIES = new Set([
  // 'Trace',
  // 'Debug',
  'Info',
  'Warn',
  'Error',
  'Fatal',
])

const DEFAULT_SELECTED_SERVICES = new Set(['api', 'runner'])

const LS_KEY_SEVERITIES = 'nuon:log-filter:severities'
const LS_KEY_SERVICES = 'nuon:log-filter:services'
const LS_KEY_JOB_OUTPUT = 'nuon:log-filter:job-output'

function readBoolFromStorage(key: string, fallback: boolean): boolean {
  try {
    const raw = localStorage.getItem(key)
    if (raw !== null) return raw === 'true'
  } catch {
    // ignore
  }
  return fallback
}

function readSetFromStorage(key: string, fallback: Set<string>): Set<string> {
  try {
    const raw = localStorage.getItem(key)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed) && parsed.length > 0) {
        return new Set(parsed)
      }
    }
  } catch {
    // ignore
  }
  return new Set(fallback)
}

export const useLogFilters = <T extends TOTELLog>(logs: T[] | null) => {
  const [selectedSeverities, setSelectedSeverities] = useState<Set<string>>(
    () => readSetFromStorage(LS_KEY_SEVERITIES, DEFAULT_SELECTED_SEVERITIES)
  )
  const [selectedServices, setSelectedServices] = useState<Set<string>>(
    () => readSetFromStorage(LS_KEY_SERVICES, DEFAULT_SELECTED_SERVICES)
  )

  useEffect(() => {
    try {
      localStorage.setItem(LS_KEY_SEVERITIES, JSON.stringify(Array.from(selectedSeverities)))
    } catch {
      // ignore
    }
  }, [selectedSeverities])

  useEffect(() => {
    try {
      localStorage.setItem(LS_KEY_SERVICES, JSON.stringify(Array.from(selectedServices)))
    } catch {
      // ignore
    }
  }, [selectedServices])
  const [jobOutputOnly, setJobOutputOnly] = useState<boolean>(
    () => readBoolFromStorage(LS_KEY_JOB_OUTPUT, false)
  )

  useEffect(() => {
    try {
      localStorage.setItem(LS_KEY_JOB_OUTPUT, String(jobOutputOnly))
    } catch {
      // ignore
    }
  }, [jobOutputOnly])

  const [searchQuery, setSearchQuery] = useState<string>('')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')

  // Get unique service names from logs
  const availableServices = useMemo(() => {
    if (!logs) return new Set<string>()

    const services = new Set<string>()
    logs.forEach((log) => {
      if (log.service_name) {
        services.add(log.service_name)
      }
    })
    return services
  }, [logs])

  // Sort logs by timestamp using full nanosecond precision
  const sortLogsByTimestamp = (logs: T[], direction: SortDirection): T[] => {
    return [...logs].sort((a, b) => {
      // Use string comparison for full nanosecond precision
      // ISO 8601 timestamps are lexicographically sortable
      const aTimestamp = a.timestamp
      const bTimestamp = b.timestamp

      if (direction === 'desc') {
        return bTimestamp > aTimestamp ? 1 : bTimestamp < aTimestamp ? -1 : 0
      } else {
        return aTimestamp > bTimestamp ? 1 : aTimestamp < bTimestamp ? -1 : 0
      }
    })
  }

  // Filter and sort logs
  const filteredLogs = useMemo(() => {
    if (!logs) return null

    // First filter by selected severities
    let filtered = logs.filter((item) =>
      selectedSeverities.has(item.severity_text)
    )

    filtered = filtered.filter((item) =>
      item.service_name ? selectedServices.has(item.service_name) : false
    )

    if (jobOutputOnly) {
      filtered = filtered.filter((item) => item.scope_name === 'oteljob')
    }

    // Then filter by search query
    if (searchQuery.trim()) {
      const searchLower = searchQuery.toLowerCase().trim()
      filtered = filtered.filter((item) =>
        item.body?.toLowerCase().includes(searchLower)
      )
    }

    // Finally sort by timestamp
    return sortLogsByTimestamp(filtered, sortDirection)
  }, [logs, selectedSeverities, selectedServices, jobOutputOnly, searchQuery, sortDirection])

  // Severity filter handlers
  const handleSeverityInputToggle = (severity: string) => {
    setSelectedSeverities((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(severity)) {
        newSet.delete(severity)
      } else {
        newSet.add(severity)
      }
      return newSet
    })
  }

  const handleSeverityButtonClick = (severity: string) => {
    setSelectedSeverities((prev) => {
      // If only this severity is selected, reset to default selected severities
      if (prev.size === 1 && prev.has(severity)) {
        return new Set(DEFAULT_SELECTED_SEVERITIES)
      }
      // Otherwise, select only this severity
      return new Set([severity])
    })
  }

  const handleSeverityReset = () => {
    setSelectedSeverities(new Set(DEFAULT_SELECTED_SEVERITIES))
  }

  // Service filter handlers
  const handleServiceInputToggle = (service: string) => {
    setSelectedServices((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(service)) {
        newSet.delete(service)
      } else {
        newSet.add(service)
      }
      return newSet
    })
  }

  const handleServiceButtonClick = (service: string) => {
    setSelectedServices((prev) => {
      // If only this service is selected, reset to all services (empty set)
      if (prev.size === 1 && prev.has(service)) {
        return new Set()
      }
      // Otherwise, select only this service
      return new Set([service])
    })
  }

  const handleServiceReset = () => {
    setSelectedServices(new Set(DEFAULT_SELECTED_SERVICES))
  }

  // Search and sort handlers
  const handleSearchChange = (query: string) => {
    setSearchQuery(query)
  }

  const handleSortToggle = () => {
    setSortDirection((prev) => (prev === 'desc' ? 'asc' : 'desc'))
  }

  const handleSortChange = (direction: SortDirection) => {
    setSortDirection(direction)
  }

  const handleJobOutputToggle = () => {
    setJobOutputOnly((prev) => !prev)
  }

  return {
    // Severity filter
    selectedSeverities,
    handleSeverityInputToggle,
    handleSeverityButtonClick,
    handleSeverityReset,

    // Service filter
    selectedServices,
    availableServices,
    handleServiceInputToggle,
    handleServiceButtonClick,
    handleServiceReset,

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
      totalCount: DEFAULT_SELECTED_SEVERITIES.size,
    },
    serviceStats: {
      selectedCount: selectedServices.size,
      totalCount: availableServices.size,
      isAllSelected: selectedServices.size === 0, // Empty set means all
    },
  }
}

export type TLogFiltersProps = ReturnType<typeof useLogFilters>

'use client'

import { useMemo, useState } from 'react'
import type { TOTELLog } from '@/types'

type SortDirection = 'asc' | 'desc'

const DEFAULT_SELECTED_SEVERITIES = new Set([
  'Trace',
  'Debug',
  'Info',
  'Warn',
  'Error',
  'Fatal',
])

const DEFAULT_SELECTED_SERVICES = new Set(['api', 'runner'])

export const useLogFilters = <T extends TOTELLog>(logs: T[] | null) => {
  const [selectedSeverities, setSelectedSeverities] = useState<Set<string>>(
    new Set(DEFAULT_SELECTED_SEVERITIES)
  )
  const [selectedServices, setSelectedServices] = useState<Set<string>>(
    new Set(DEFAULT_SELECTED_SERVICES) // Start with all services selected (empty set means all)
  )
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

  // Sort logs by timestamp
  const sortLogsByTimestamp = (logs: T[], direction: SortDirection): T[] => {
    return [...logs].sort((a, b) => {
      const aTime = new Date(a.timestamp).getTime()
      const bTime = new Date(b.timestamp).getTime()

      return direction === 'desc' ? bTime - aTime : aTime - bTime
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

    // Then filter by search query
    if (searchQuery.trim()) {
      const searchLower = searchQuery.toLowerCase().trim()
      filtered = filtered.filter((item) =>
        item.body?.toLowerCase().includes(searchLower)
      )
    }

    // Finally sort by timestamp
    return sortLogsByTimestamp(filtered, sortDirection)
  }, [logs, selectedSeverities, selectedServices, searchQuery, sortDirection])

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

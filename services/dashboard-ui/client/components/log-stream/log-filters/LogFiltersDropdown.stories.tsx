export default {
  title: 'LogStream/LogFiltersDropdown',
}

import { useState } from 'react'
import { LogFiltersDropdown } from './LogFiltersDropdown'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

const noop = () => {}

const useMockFilters = ({
  tools = [] as string[],
  systemLogs = false,
  sortDir = 'desc' as 'asc' | 'desc',
} = {}) => {
  const [tool, setTool] = useState('')
  const [includeSystemLogs, setIncludeSystemLogs] = useState(systemLogs)
  const [sort, setSort] = useState(sortDir)

  const isFiltered = includeSystemLogs || !!tool

  return {
    includeSystemLogs,
    handleSystemLogsToggle: () => setIncludeSystemLogs((v) => !v),
    tool,
    setTool,
    availableTools: new Set(tools),
    sortStats: {
      direction: sort,
      isNewestFirst: sort === 'desc',
      isOldestFirst: sort === 'asc',
    },
    handleSortToggle: () => setSort((s) => (s === 'desc' ? 'asc' : 'desc')),
    isFiltered,
    handleResetAll: () => {
      setTool('')
      setIncludeSystemLogs(false)
      setSort('desc')
    },
  } as unknown as TLogFiltersProps
}

export const Default = () => {
  const filters = useMockFilters()
  return <LogFiltersDropdown filters={filters} />
}

export const WithTools = () => {
  const filters = useMockFilters({ tools: ['helm', 'terraform', 'kustomize'] })
  return <LogFiltersDropdown filters={filters} />
}

export const SystemLogsEnabled = () => {
  const filters = useMockFilters({
    tools: ['helm', 'terraform'],
    systemLogs: true,
  })
  return <LogFiltersDropdown filters={filters} />
}

export const OldestFirst = () => {
  const filters = useMockFilters({ sortDir: 'asc' })
  return <LogFiltersDropdown filters={filters} />
}

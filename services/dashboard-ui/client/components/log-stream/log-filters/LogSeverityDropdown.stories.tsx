export default {
  title: 'LogStream/LogSeverityDropdown',
}

import { useState } from 'react'
import { LogSeverityDropdown } from './LogSeverityDropdown'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

const KNOWN = ['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal']

const useMockSeverityFilters = (initial: string[]) => {
  const [selected, setSelected] = useState(new Set(initial))

  const handleSeverityInputToggle = (s: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(s)) next.delete(s)
      else next.add(s)
      return next
    })
  }

  const handleSeverityButtonClick = (s: string) => {
    setSelected((prev) => {
      if (prev.size === 1 && prev.has(s)) return new Set(KNOWN)
      return new Set([s])
    })
  }

  const handleSeverityReset = () => setSelected(new Set(KNOWN))

  return {
    selectedSeverities: selected,
    availableSeverities: new Set(KNOWN),
    handleSeverityInputToggle,
    handleSeverityButtonClick,
    handleSeverityReset,
    severityStats: {
      selectedCount: selected.size,
      totalCount: KNOWN.length,
      isDefault: selected.size === 4 && !selected.has('Trace') && !selected.has('Debug'),
    },
  } as unknown as TLogFiltersProps
}

export const Default = () => {
  const filters = useMockSeverityFilters(['Info', 'Warn', 'Error', 'Fatal'])
  return <LogSeverityDropdown filters={filters} />
}

export const AllSelected = () => {
  const filters = useMockSeverityFilters(KNOWN)
  return <LogSeverityDropdown filters={filters} />
}

export const SingleSelected = () => {
  const filters = useMockSeverityFilters(['Error'])
  return <LogSeverityDropdown filters={filters} />
}

export const NoneSelected = () => {
  const filters = useMockSeverityFilters([])
  return <LogSeverityDropdown filters={filters} />
}

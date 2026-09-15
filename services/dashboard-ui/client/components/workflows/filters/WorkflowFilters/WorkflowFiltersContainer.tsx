import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import {
  readWorkflowFilters,
  type TWorkflowDatePreset,
  type TWorkflowOwner,
  type TWorkflowPreviewOption,
  type TWorkflowStatusOption,
  type TWorkflowTypeGroup,
} from '@/utils/workflow-filters'
import { WorkflowFilters } from './WorkflowFilters'

const SEARCH_DEBOUNCE_MS = 300

export interface IWorkflowFiltersContainer {
  owner: TWorkflowOwner
}

export const WorkflowFiltersContainer = ({
  owner,
}: IWorkflowFiltersContainer) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const filters = readWorkflowFilters(searchParams, owner)
  const [search, setSearch] = useState(filters.search)
  const searchTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    setSearch(filters.search)
  }, [filters.search])

  useEffect(
    () => () => {
      if (searchTimer.current) clearTimeout(searchTimer.current)
    },
    []
  )

  const writeSet = (key: string, selected: Set<string>) => {
    setSearchParams(
      (current) => {
        const next = new URLSearchParams(current)
        if (selected.size) next.set(key, [...selected].join(','))
        else next.delete(key)
        next.delete('offset')
        return next
      },
      { replace: true }
    )
  }

  const handleSearchChange = (value: string) => {
    setSearch(value)
    if (searchTimer.current) clearTimeout(searchTimer.current)
    searchTimer.current = setTimeout(() => {
      setSearchParams(
        (current) => {
          const next = new URLSearchParams(current)
          if (value) next.set('q', value)
          else next.delete('q')
          next.delete('search')
          next.delete('offset')
          return next
        },
        { replace: true }
      )
    }, SEARCH_DEBOUNCE_MS)
  }

  return (
    <WorkflowFilters
      owner={owner}
      search={search}
      status={filters.status}
      type={filters.type}
      date={filters.date}
      preview={filters.preview}
      onSearchChange={handleSearchChange}
      onStatusChange={(value: Set<TWorkflowStatusOption>) =>
        writeSet('status', value)
      }
      onTypeChange={(value: Set<TWorkflowTypeGroup>) => writeSet('type', value)}
      onDateChange={(value: Set<TWorkflowDatePreset>) =>
        writeSet('since', value)
      }
      onPreviewChange={(value: Set<TWorkflowPreviewOption>) =>
        writeSet('preview', value)
      }
    />
  )
}

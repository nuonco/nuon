import { useEffect, useMemo, useState } from 'react'
import type { IPlanDiffSection, TDiffOperation } from '../lib/diffs'
import { useFilterSelection } from './use-filter-selection'

export const usePlanDiffFilter = (
  sections: IPlanDiffSection[],
  availableOperations?: readonly TDiffOperation[],
  defaultOperations?: readonly TDiffOperation[]
) => {
  const operations = useMemo(
    () => [
      ...new Set(
        availableOperations ?? sections.map(({ operation }) => operation)
      ),
    ],
    [availableOperations, sections]
  )
  const selection = useFilterSelection(operations, defaultOperations)
  const [searchQuery, setSearchQuery] = useState('')

  useEffect(() => {
    setSearchQuery('')
  }, [operations])

  const filteredSections = useMemo(() => {
    const query = searchQuery.trim().toLowerCase()

    return sections.filter(
      ({ operation, searchable }) =>
        selection.selected.has(operation) &&
        (!query ||
          searchable.some((value) => value.toLowerCase().includes(query)))
    )
  }, [searchQuery, sections, selection.selected])

  const reset = () => {
    selection.reset()
    setSearchQuery('')
  }

  return {
    filteredSections,
    operations,
    reset,
    searchQuery,
    selectedCount: filteredSections.length,
    selectedOperations: selection.selected,
    setSearchQuery,
    toggleOperation: selection.toggle,
    onlyOperation: selection.isolate,
    totalCount: sections.length,
  }
}

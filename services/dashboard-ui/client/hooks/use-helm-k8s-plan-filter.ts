import { useMemo, useState } from 'react'

const DEFAULT_SELECTED_ACTIONS = new Set(['added', 'destroyed', 'changed'])

export interface HelmFilterableItem {
  action: string
  release?: string
  resource?: string
  resourceType?: string
  workspace?: string
}

export const useHelmK8sPlanFilter = <T extends HelmFilterableItem>(
  items: T[] | null
) => {
  const [selectedActions, setSelectedActions] = useState<Set<string>>(
    new Set(DEFAULT_SELECTED_ACTIONS)
  )
  const [searchQuery, setSearchQuery] = useState<string>('')

  const filteredItems = useMemo(() => {
    if (!items) return null

    let filtered = items.filter((item) => selectedActions.has(item.action))

    if (searchQuery.trim()) {
      const searchLower = searchQuery.toLowerCase().trim()
      filtered = filtered.filter(
        (item) =>
          item.release?.toLowerCase().includes(searchLower) ||
          item.resource?.toLowerCase().includes(searchLower) ||
          item.resourceType?.toLowerCase().includes(searchLower) ||
          item.workspace?.toLowerCase().includes(searchLower)
      )
    }

    return filtered
  }, [items, selectedActions, searchQuery])

  const handleInputToggle = (action: string) => {
    setSelectedActions((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(action)) {
        newSet.delete(action)
      } else {
        newSet.add(action)
      }
      return newSet
    })
  }

  const handleButtonClick = (action: string) => {
    setSelectedActions((prev) => {
      if (prev.size === 1 && prev.has(action)) {
        return new Set(DEFAULT_SELECTED_ACTIONS)
      }
      return new Set([action])
    })
  }

  const handleReset = () => {
    setSelectedActions(new Set(DEFAULT_SELECTED_ACTIONS))
  }

  const handleSearchChange = (query: string) => {
    setSearchQuery(query)
  }

  return {
    selectedActions,
    searchQuery,
    filteredItems,
    handleInputToggle,
    handleButtonClick,
    handleReset,
    handleSearchChange,
    filterStats: {
      selectedCount: filteredItems?.length || 0,
      totalCount: items?.length || 0,
    },
  }
}

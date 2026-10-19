import { useMemo, useState } from 'react'

// Default selected actions (all except read and noop)
const DEFAULT_SELECTED_ACTIONS = new Set([
  'create',
  'update',
  'delete',
  'replace',
])

export interface TerraformResourceFilterableItem {
  action: string
  address?: string
  resource?: string
  name?: string
}

export const useTerraformResourceFilter = <
  T extends TerraformResourceFilterableItem,
>(
  items: T[]
) => {
  const [selectedActions, setSelectedActions] = useState<Set<string>>(
    new Set(DEFAULT_SELECTED_ACTIONS)
  )
  const [searchQuery, setSearchQuery] = useState<string>('')

  const filteredItems = useMemo(() => {
    let filtered = items.filter((item) => selectedActions.has(item.action))

    if (searchQuery.trim()) {
      const searchLower = searchQuery.toLowerCase().trim()
      filtered = filtered.filter(
        (item) =>
          item.address?.toLowerCase().includes(searchLower) ||
          item.resource?.toLowerCase().includes(searchLower) ||
          item.name?.toLowerCase().includes(searchLower)
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
      selectedCount: filteredItems.length,
      totalCount: items.length,
    },
  }
}

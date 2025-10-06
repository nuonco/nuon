'use client'

import { useMemo, useState } from 'react'

// Default selected actions (all except read and noop)
const DEFAULT_SELECTED_ACTIONS = new Set([
  'create',
  'update',
  'delete',
  'replace',
])

export interface TerraformOutputFilterableItem {
  action: string
  output?: string
}

export const useTerraformOutputFilter = <
  T extends TerraformOutputFilterableItem,
>(
  items: T[]
) => {
  const [selectedActions, setSelectedActions] = useState<Set<string>>(
    new Set(DEFAULT_SELECTED_ACTIONS)
  )
  const [searchQuery, setSearchQuery] = useState<string>('')

  // Filter items based on selected actions and search query
  const filteredItems = useMemo(() => {
    // First filter by selected actions
    let filtered = items.filter((item) => selectedActions.has(item.action))

    // Then filter by search query (search by output only)
    if (searchQuery.trim()) {
      const searchLower = searchQuery.toLowerCase().trim()
      filtered = filtered.filter((item) =>
        item.output?.toLowerCase().includes(searchLower)
      )
    }

    return filtered
  }, [items, selectedActions, searchQuery])

  // Checkbox behavior: toggle the action
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

  // Button behavior: select only this action OR reset if only this action is selected
  const handleButtonClick = (action: string) => {
    setSelectedActions((prev) => {
      // If only this action is selected, reset to default selected actions
      if (prev.size === 1 && prev.has(action)) {
        return new Set(DEFAULT_SELECTED_ACTIONS)
      }
      // Otherwise, select only this action
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

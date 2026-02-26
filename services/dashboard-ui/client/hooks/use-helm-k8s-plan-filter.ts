'use client'

import { useMemo, useState } from 'react'

// Default selected actions for Helm (all actions selected by default)
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

  // Filter items based on selected actions and search query
  const filteredItems = useMemo(() => {
    if (!items) return null

    let filtered = items.filter((item) => selectedActions.has(item.action))

    // Filter by search query (search by release, resource, or resourceType)
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

  // Action handlers
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
      selectedCount: filteredItems?.length || 0,
      totalCount: items?.length || 0,
    },
  }
}

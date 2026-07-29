import { useMemo, useState } from 'react'

export type TAttributeFilter = 'required' | 'sensitive'
export type TSourceFilter = 'vendor' | 'customer'

export const ATTRIBUTE_OPTIONS: TAttributeFilter[] = ['required', 'sensitive']
export const SOURCE_OPTIONS: TSourceFilter[] = ['vendor', 'customer']

export const ATTRIBUTE_LABELS: Record<TAttributeFilter, string> = {
  required: 'Required',
  sensitive: 'Sensitive',
}

export const SOURCE_LABELS: Record<TSourceFilter, string> = {
  vendor: 'Vendor',
  customer: 'Customer',
}

export type TInputsFilterInput = {
  name?: string
  display_name?: string
  description?: string
  default?: string
  required?: boolean
  sensitive?: boolean
  source?: string
  group_id?: string
}

export type TInputsFilterGroup = {
  id: string
  name?: string
  display_name?: string
  description?: string
  app_inputs?: TInputsFilterInput[]
}

export const useInputsFilter = ({
  inputGroups,
  redacted,
}: {
  inputGroups: TInputsFilterGroup[]
  redacted: Record<string, any>
}) => {
  const [search, setSearch] = useState('')
  const [attributeFilters, setAttributeFilters] = useState<TAttributeFilter[]>(
    []
  )
  const [sourceFilters, setSourceFilters] = useState<TSourceFilter[]>([])

  const query = search.toLowerCase()
  const hasActiveSearch = query.length > 0
  const filterCount = attributeFilters.length + sourceFilters.length
  const hasActiveFilters = filterCount > 0

  const clearAllFilters = () => {
    setAttributeFilters([])
    setSourceFilters([])
  }

  const clearAll = () => {
    setSearch('')
    clearAllFilters()
  }

  const toggleAttribute = (filter: TAttributeFilter) => {
    setAttributeFilters((prev) =>
      prev.includes(filter)
        ? prev.filter((f) => f !== filter)
        : [...prev, filter]
    )
  }

  const toggleSource = (filter: TSourceFilter) => {
    setSourceFilters((prev) =>
      prev.includes(filter)
        ? prev.filter((f) => f !== filter)
        : [...prev, filter]
    )
  }

  const matchesFilters = (input: TInputsFilterInput) => {
    if (!hasActiveFilters) return true

    const matchesAttributes = attributeFilters.every((f) => {
      if (f === 'required') return input.required
      if (f === 'sensitive') return input.sensitive
      return false
    })

    const matchesSource =
      sourceFilters.length === 0 ||
      sourceFilters.some((f) => input.source === f)

    return matchesAttributes && matchesSource
  }

  const filteredGroups = useMemo(() => {
    return inputGroups
      .map((group) => ({
        ...group,
        app_inputs: (group.app_inputs ?? []).filter((input) => {
          if (!matchesFilters(input)) return false
          if (!query) return true
          const val = input.name ? String(redacted[input.name] ?? '') : ''
          return (
            input.name?.toLowerCase().includes(query) ||
            input.display_name?.toLowerCase().includes(query) ||
            input.description?.toLowerCase().includes(query) ||
            val.toLowerCase().includes(query)
          )
        }),
      }))
      .filter((group) => (group.app_inputs?.length ?? 0) > 0)
  }, [query, inputGroups, redacted, attributeFilters, sourceFilters])

  const filteredFlatInputs = useMemo(() => {
    if (!query) return Object.entries(redacted)
    return Object.entries(redacted).filter(
      ([key, value]) =>
        key.toLowerCase().includes(query) ||
        String(value).toLowerCase().includes(query)
    )
  }, [query, redacted])

  return {
    search,
    setSearch,
    attributeFilters,
    sourceFilters,
    setAttributeFilters,
    setSourceFilters,
    toggleAttribute,
    toggleSource,
    clearAllFilters,
    clearAll,
    filterCount,
    hasActiveSearch,
    hasActiveFilters,
    filteredGroups,
    filteredFlatInputs,
  }
}

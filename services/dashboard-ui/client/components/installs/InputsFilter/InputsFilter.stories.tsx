import { useState } from 'react'
import type {
  TAttributeFilter,
  TSourceFilter,
} from '@/hooks/use-inputs-filter'
import { InputsFilterBar } from './InputsFilterBar'
import { InputsNoResults } from './InputsNoResults'

export default {
  title: 'Installs/InputsFilter',
}

export const FilterBar = () => {
  const [search, setSearch] = useState('')
  const [attributeFilters, setAttributeFilters] = useState<TAttributeFilter[]>(
    []
  )
  const [sourceFilters, setSourceFilters] = useState<TSourceFilter[]>([])
  const filterCount = attributeFilters.length + sourceFilters.length

  return (
    <InputsFilterBar
      search={search}
      onSearchChange={setSearch}
      showFilters
      attributeFilters={attributeFilters}
      sourceFilters={sourceFilters}
      setAttributeFilters={setAttributeFilters}
      setSourceFilters={setSourceFilters}
      toggleAttribute={(f) =>
        setAttributeFilters((prev) =>
          prev.includes(f) ? prev.filter((x) => x !== f) : [...prev, f]
        )
      }
      toggleSource={(f) =>
        setSourceFilters((prev) =>
          prev.includes(f) ? prev.filter((x) => x !== f) : [...prev, f]
        )
      }
      clearAllFilters={() => {
        setAttributeFilters([])
        setSourceFilters([])
      }}
      filterCount={filterCount}
      hasActiveFilters={filterCount > 0}
    />
  )
}

export const FilterBarSearchOnly = () => {
  const [search, setSearch] = useState('')

  return (
    <InputsFilterBar
      search={search}
      onSearchChange={setSearch}
      showFilters={false}
      attributeFilters={[]}
      sourceFilters={[]}
      setAttributeFilters={() => {}}
      setSourceFilters={() => {}}
      toggleAttribute={() => {}}
      toggleSource={() => {}}
      clearAllFilters={() => {}}
      filterCount={0}
      hasActiveFilters={false}
    />
  )
}

export const NoResultsSearch = () => (
  <InputsNoResults
    search="database"
    hasActiveSearch
    hasActiveFilters={false}
    onClearSearch={() => {}}
    onClearFilters={() => {}}
    onClearAll={() => {}}
  />
)

export const NoResultsSearchAndFilters = () => (
  <InputsNoResults
    search="database"
    hasActiveSearch
    hasActiveFilters
    onClearSearch={() => {}}
    onClearFilters={() => {}}
    onClearAll={() => {}}
  />
)

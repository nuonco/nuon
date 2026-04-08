export default {
  title: 'LogStream/LogSearch',
}

import { useState } from 'react'
import { LogSearch } from './LogSearch'

export const Default = () => {
  const [searchQuery, setSearchQuery] = useState('')
  return (
    <LogSearch
      filters={{
        searchQuery,
        handleSearchChange: setSearchQuery,
        filterStats: { selectedCount: 42, totalCount: 100 },
      }}
    />
  )
}

export const WithQuery = () => (
  <LogSearch
    filters={{
      searchQuery: 'error',
      handleSearchChange: () => {},
      filterStats: { selectedCount: 3, totalCount: 100 },
    }}
  />
)

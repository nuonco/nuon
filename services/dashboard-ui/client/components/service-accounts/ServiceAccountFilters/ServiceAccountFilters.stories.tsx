export default {
  title: 'Features / Service accounts / Service account filters',
}

import { useState } from 'react'
import type { TServiceAccountManagement } from '@/types'
import { ServiceAccountFilters } from './ServiceAccountFilters'

const FiltersStory = ({
  initialManagement = 'user',
  initialPurpose = '',
  initialSearch = '',
  purposes = ['stack', 'telemetry'],
}: {
  initialManagement?: TServiceAccountManagement
  initialPurpose?: string
  initialSearch?: string
  purposes?: string[]
}) => {
  const [management, setManagement] = useState(initialManagement)
  const [purpose, setPurpose] = useState(initialPurpose)
  const [search, setSearch] = useState(initialSearch)

  return (
    <ServiceAccountFilters
      management={management}
      purpose={purpose}
      purposes={purposes}
      search={search}
      onManagementChange={(value) => {
        setManagement(value)
        if (value === 'user') setPurpose('')
      }}
      onPurposeChange={setPurpose}
      onSearchChange={setSearch}
    />
  )
}

export const Default = () => <FiltersStory />

export const SystemWithPurpose = () => (
  <FiltersStory initialManagement="system" initialPurpose="stack" />
)

export const WithSearch = () => (
  <FiltersStory initialManagement="all" initialSearch="acme" />
)

export const NoPurposes = () => (
  <FiltersStory initialManagement="system" purposes={[]} />
)

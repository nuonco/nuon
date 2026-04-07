export default {
  title: 'Policies/PolicyReportsFilter',
}

import { PolicyReportsFilter } from './PolicyReportsFilter'

export const Default = () => (
  <div className="p-4">
    <PolicyReportsFilter
      onStatusChange={() => {}}
      onTypeChange={() => {}}
      onClearFilters={() => {}}
    />
  </div>
)

export const WithFilters = () => (
  <div className="p-4">
    <PolicyReportsFilter
      currentStatus="error"
      currentOwnerType="install_deploys"
      onStatusChange={() => {}}
      onTypeChange={() => {}}
      onClearFilters={() => {}}
    />
  </div>
)

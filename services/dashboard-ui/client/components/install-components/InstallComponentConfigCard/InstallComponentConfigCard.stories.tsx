export default {
  title: 'Install Components/InstallComponentConfigCard',
}

import { InstallComponentConfigCard, InstallComponentConfigCardSkeleton } from './InstallComponentConfigCard'

const mockConfig = {
  component_id: 'comp-1',
  version: 3,
  type: 'helm_chart',
  helm: {
    chart_name: 'my-chart',
    namespace: 'production',
  },
} as any

export const Default = () => (
  <InstallComponentConfigCard
    config={mockConfig}
    orgId="org-1"
    installAppId="app-1"
    installAppConfigId="config-1"
  />
)

export const Loading = () => <InstallComponentConfigCardSkeleton />

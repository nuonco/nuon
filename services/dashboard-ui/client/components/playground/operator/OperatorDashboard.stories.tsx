import { OperatorDashboard } from './OperatorDashboard'
import { operatorInstalls } from './fixtures'

export default {
  title: 'Playground/Operator/Dashboard',
}

export const Default = () => <OperatorDashboard />

export const OutOfDate = () => <OperatorDashboard initialFilter="behind" />

export const NeedsAttention = () => (
  <OperatorDashboard initialFilter="attention" />
)

export const AllCurrent = () => (
  <OperatorDashboard
    installs={operatorInstalls.map((install) => ({
      ...install,
      app_config_version: install.app_config_latest_version,
      deployments_status: 'active',
      deployments_detail: 'All components deployed',
      resources_status: 'ok',
      resources_detail: 'No drift detected',
      health_status: 'healthy',
      health_detail: 'All components passing health checks',
    }))}
  />
)

export const Empty = () => <OperatorDashboard installs={[]} />

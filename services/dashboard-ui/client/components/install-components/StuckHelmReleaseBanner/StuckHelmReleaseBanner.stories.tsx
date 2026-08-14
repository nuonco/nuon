export default {
  title: 'InstallComponents/StuckHelmReleaseBanner',
}

import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TComponent } from '@/types'
import { StuckHelmReleaseBanner } from './StuckHelmReleaseBanner'

const component = {
  id: 'cmp000000000000000000000001',
  name: 'api-server',
  type: 'helm_chart',
} as TComponent

export const PendingUpgrade = () => (
  <SurfacesProvider>
    <StuckHelmReleaseBanner component={component} status="pending-upgrade" />
  </SurfacesProvider>
)

export const StatusUnknown = () => (
  <SurfacesProvider>
    <StuckHelmReleaseBanner component={component} />
  </SurfacesProvider>
)

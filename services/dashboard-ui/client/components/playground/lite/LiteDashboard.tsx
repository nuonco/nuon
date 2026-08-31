import { AppRoutes } from './AppRoutes'
import { Shell } from './Shell'
import { primaryNav, secondaryNav } from './nav'

export const LiteDashboard = () => (
  <Shell primaryNav={primaryNav} secondaryNav={secondaryNav}>
    <AppRoutes />
  </Shell>
)

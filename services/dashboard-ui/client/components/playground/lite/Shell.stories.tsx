import { AppRoutes } from './AppRoutes'
import { Page } from './Page'
import { PlaceholderGrid } from './PlaceholderGrid'
import { Shell } from './Shell'
import { primaryNav, secondaryNav } from './nav'

export default {
  title: 'Playground/Lite/Shell',
}

export const Default = () => (
  <Shell primaryNav={primaryNav} secondaryNav={secondaryNav}>
    <AppRoutes />
  </Shell>
)
Default.meta = { fullBleed: true }

export const PrimaryNavOnly = () => (
  <Shell primaryNav={primaryNav}>
    <AppRoutes />
  </Shell>
)
PrimaryNavOnly.meta = { fullBleed: true }

export const StaticContent = () => (
  <Shell primaryNav={primaryNav} secondaryNav={secondaryNav}>
    <Page crumbs={[{ label: 'Home' }]}>
      <PlaceholderGrid rows={1} height="h-[12rem]" />
    </Page>
  </Shell>
)
StaticContent.meta = { fullBleed: true }

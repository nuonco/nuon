import { TabNav } from './TabNav'
import { appTabs } from './nav'

export default {
  title: 'Playground/Lite/TabNav',
}

export const Default = () => <TabNav tabs={appTabs('app-01', 'br-main')} />

export const Many = () => (
  <TabNav
    tabs={[
      ...appTabs('app-01', 'br-main'),
      { label: 'Releases', path: '/apps/app-01/branches/br-main/releases' },
      { label: 'Settings', path: '/apps/app-01/branches/br-main/settings' },
    ]}
  />
)

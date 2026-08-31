import { Outlet, useLocation } from 'react-router'
import { Page } from './Page'
import { settingsNav } from './nav'

export const SettingsLayout = () => {
  const { pathname } = useLocation()
  const active = settingsNav.reduce((match, item) =>
    pathname.startsWith(item.path) && item.path.length > match.path.length
      ? item
      : match
  )

  return (
    <Page
      tabs={settingsNav}
      crumbs={[
        { label: 'Settings', path: '/settings' },
        { label: active.label },
      ]}
    >
      <Outlet />
    </Page>
  )
}

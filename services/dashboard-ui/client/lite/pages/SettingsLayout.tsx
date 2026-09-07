import { Outlet } from 'react-router'
import { SubNav, type ISubNavItem } from '../components/molecules/SubNav'
import { useOrg } from '../providers/org-provider'

export const settingsNavigation = (orgId: string): ISubNavItem[] => {
  const base = `/${orgId}/settings`

  return [
    { href: base, label: 'Connections', end: true },
    { href: `${base}/webhooks`, label: 'Webhooks' },
    { href: `${base}/triggers`, label: 'Triggers' },
    { href: `${base}/api-tokens`, label: 'API tokens' },
    { href: `${base}/service-accounts`, label: 'Service accounts' },
    { href: `${base}/oidc`, label: 'OIDC federation' },
  ]
}

export const SettingsLayout = () => {
  const { orgId } = useOrg()

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <SubNav
        items={settingsNavigation(orgId ?? '')}
        label="Settings sections"
      />
      <Outlet />
    </div>
  )
}

import { Card } from '../components/atoms/Card'
import { Text } from '../components/atoms/Text'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import type { IBreadcrumbItem } from '../providers/breadcrumb-provider'
import { useOrg } from '../providers/org-provider'

const useOrgPageChrome = ({
  label,
  orgRoot = false,
  settings = false,
}: {
  label: string
  orgRoot?: boolean
  settings?: boolean
}) => {
  const { org, orgId } = useOrg()

  const trail: IBreadcrumbItem[] = [
    {
      label: org?.name,
      href: orgRoot || !orgId ? undefined : `/${orgId}`,
      loadingWidth: 16,
    },
  ]
  if (settings) {
    trail.push({
      label: 'Settings',
      href: orgId ? `/${orgId}/settings` : undefined,
    })
  }
  if (!orgRoot) trail.push({ label })

  usePageTitle(label)
  useBreadcrumbs(trail)
}

const Placeholder = () => (
  <Card className="min-h-40">
    <Text variant="caption" color="tertiary">
      Page content will be added in a follow-up.
    </Text>
  </Card>
)

export const Dashboard = () => {
  useOrgPageChrome({ label: 'Dashboard', orgRoot: true })
  return <Placeholder />
}

export const Teams = () => {
  useOrgPageChrome({ label: 'Team' })
  return <Placeholder />
}

export const Connections = () => {
  useOrgPageChrome({ label: 'Connections', settings: true })
  return <Placeholder />
}

export const Webhooks = () => {
  useOrgPageChrome({ label: 'Webhooks', settings: true })
  return <Placeholder />
}

export const Triggers = () => {
  useOrgPageChrome({ label: 'Triggers', settings: true })
  return <Placeholder />
}

export const ApiTokens = () => {
  useOrgPageChrome({ label: 'API tokens', settings: true })
  return <Placeholder />
}

export const ServiceAccounts = () => {
  useOrgPageChrome({ label: 'Service accounts', settings: true })
  return <Placeholder />
}

export const OidcFederation = () => {
  useOrgPageChrome({ label: 'OIDC federation', settings: true })
  return <Placeholder />
}

export const Onboarding = () => {
  usePageTitle('Onboarding')
  return (
    <div className="flex flex-col gap-1">
      <Text as="h1" variant="title">
        Welcome to Nuon
      </Text>
      <Text as="p" variant="caption" color="secondary">
        Configure your account and first organization.
      </Text>
    </div>
  )
}

export const NotFound = () => {
  useOrgPageChrome({ label: 'Page not found' })
  return <Placeholder />
}

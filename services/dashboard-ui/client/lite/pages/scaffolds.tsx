import { RouteScaffold } from '../components/organisms/RouteScaffold'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import type { IBreadcrumbItem } from '../providers/breadcrumb-provider'
import { useOrg } from '../providers/org-provider'

const useOrgPageChrome = ({
  label,
  isOrgRoot = false,
  settings = false,
}: {
  label: string
  isOrgRoot?: boolean
  settings?: boolean
}) => {
  const { org, orgId } = useOrg()

  const trail: IBreadcrumbItem[] = [
    {
      label: org?.name,
      href: isOrgRoot || !orgId ? undefined : `/${orgId}`,
      loadingWidth: 16,
    },
  ]
  if (settings) {
    trail.push({
      label: 'Settings',
      href: orgId ? `/${orgId}/settings` : undefined,
    })
  }
  if (!isOrgRoot) trail.push({ label })

  usePageTitle(label)
  useBreadcrumbs(trail)
}

export const Dashboard = () => {
  useOrgPageChrome({ label: 'Dashboard', isOrgRoot: true })
  return (
    <RouteScaffold
      title="Dashboard"
      description="Review activity across this organization."
    />
  )
}

export const Teams = () => {
  useOrgPageChrome({ label: 'Team' })
  return (
    <RouteScaffold
      title="Team"
      description="Manage organization members and access."
    />
  )
}

export const Connections = () => {
  useOrgPageChrome({ label: 'Connections', settings: true })
  return (
    <RouteScaffold
      title="Connections"
      description="Manage GitHub and Slack connections."
    />
  )
}

export const Webhooks = () => {
  useOrgPageChrome({ label: 'Webhooks', settings: true })
  return (
    <RouteScaffold
      title="Webhooks"
      description="Manage webhook destinations and subscriptions."
    />
  )
}

export const Triggers = () => {
  useOrgPageChrome({ label: 'Triggers', settings: true })
  return (
    <RouteScaffold
      title="Triggers"
      description="Manage event-driven automation."
    />
  )
}

export const ApiTokens = () => {
  useOrgPageChrome({ label: 'API tokens', settings: true })
  return (
    <RouteScaffold
      title="API tokens"
      description="Manage organization API tokens."
    />
  )
}

export const ServiceAccounts = () => {
  useOrgPageChrome({ label: 'Service accounts', settings: true })
  return (
    <RouteScaffold
      title="Service accounts"
      description="Manage non-human organization access."
    />
  )
}

export const OidcFederation = () => {
  useOrgPageChrome({ label: 'OIDC federation', settings: true })
  return (
    <RouteScaffold
      title="OIDC federation"
      description="Manage federated workload identities."
    />
  )
}

export const Onboarding = () => {
  usePageTitle('Onboarding')
  return (
    <RouteScaffold
      title="Welcome to Nuon"
      description="Configure your account and first organization."
    />
  )
}

export const NotFound = () => {
  useOrgPageChrome({ label: 'Page not found' })
  return (
    <RouteScaffold
      title="Page not found"
      description="The requested Lite dashboard page does not exist."
    />
  )
}

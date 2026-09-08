import { RouteScaffold } from '../components/organisms/RouteScaffold'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { useOrg } from '../providers/org-provider'

const useOrgPageBreadcrumbs = ({
  label,
  settings = false,
}: {
  label: string
  settings?: boolean
}) => {
  const { org, orgId } = useOrg()
  const orgHref = orgId ? `/${orgId}` : undefined
  const settingsHref = orgId ? `/${orgId}/settings` : undefined

  useBreadcrumbs([
    {
      label: org?.name,
      href: label === 'Dashboard' ? undefined : orgHref,
      loadingWidth: 16,
    },
    ...(label === 'Dashboard'
      ? []
      : settings
        ? [
            { label: 'Settings', href: settingsHref },
            { label },
          ]
        : [{ label }]),
  ])
}

export const Dashboard = () => {
  useOrgPageBreadcrumbs({ label: 'Dashboard' })
  return (
    <RouteScaffold
      title="Dashboard"
      description="Review activity across this organization."
    />
  )
}

export const Teams = () => {
  useOrgPageBreadcrumbs({ label: 'Team' })
  return (
    <RouteScaffold
      title="Team"
      description="Manage organization members and access."
    />
  )
}

export const Connections = () => {
  useOrgPageBreadcrumbs({ label: 'Connections', settings: true })
  return (
    <RouteScaffold
      title="Connections"
      description="Manage GitHub and Slack connections."
    />
  )
}

export const Webhooks = () => {
  useOrgPageBreadcrumbs({ label: 'Webhooks', settings: true })
  return (
    <RouteScaffold
      title="Webhooks"
      description="Manage webhook destinations and subscriptions."
    />
  )
}

export const Triggers = () => {
  useOrgPageBreadcrumbs({ label: 'Triggers', settings: true })
  return (
    <RouteScaffold
      title="Triggers"
      description="Manage event-driven automation."
    />
  )
}

export const ApiTokens = () => {
  useOrgPageBreadcrumbs({ label: 'API tokens', settings: true })
  return (
    <RouteScaffold
      title="API tokens"
      description="Manage organization API tokens."
    />
  )
}

export const ServiceAccounts = () => {
  useOrgPageBreadcrumbs({ label: 'Service accounts', settings: true })
  return (
    <RouteScaffold
      title="Service accounts"
      description="Manage non-human organization access."
    />
  )
}

export const OidcFederation = () => {
  useOrgPageBreadcrumbs({ label: 'OIDC federation', settings: true })
  return (
    <RouteScaffold
      title="OIDC federation"
      description="Manage federated workload identities."
    />
  )
}

export const Onboarding = () => (
  <RouteScaffold
    title="Welcome to Nuon"
    description="Configure your account and first organization."
  />
)

export const NotFound = () => {
  useOrgPageBreadcrumbs({ label: 'Page not found' })
  return (
    <RouteScaffold
      title="Page not found"
      description="The requested Lite dashboard page does not exist."
    />
  )
}

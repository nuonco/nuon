import { RouteScaffold } from '../components/organisms/RouteScaffold'

export const Dashboard = () => (
  <RouteScaffold
    title="Dashboard"
    description="Review activity across this organization."
  />
)

export const Apps = () => (
  <RouteScaffold
    title="Apps"
    description="Manage applications for this organization."
  />
)

export const Installs = () => (
  <RouteScaffold
    title="Installs"
    description="Manage installations for this organization."
  />
)

export const Teams = () => (
  <RouteScaffold
    title="Team"
    description="Manage organization members and access."
  />
)

export const Connections = () => (
  <RouteScaffold
    title="Connections"
    description="Manage GitHub and Slack connections."
  />
)

export const Webhooks = () => (
  <RouteScaffold
    title="Webhooks"
    description="Manage webhook destinations and subscriptions."
  />
)

export const Triggers = () => (
  <RouteScaffold
    title="Triggers"
    description="Manage event-driven automation."
  />
)

export const ApiTokens = () => (
  <RouteScaffold
    title="API tokens"
    description="Manage organization API tokens."
  />
)

export const ServiceAccounts = () => (
  <RouteScaffold
    title="Service accounts"
    description="Manage non-human organization access."
  />
)

export const OidcFederation = () => (
  <RouteScaffold
    title="OIDC federation"
    description="Manage federated workload identities."
  />
)

export const Onboarding = () => (
  <RouteScaffold
    title="Welcome to Nuon"
    description="Configure your account and first organization."
  />
)

export const NotFound = () => (
  <RouteScaffold
    title="Page not found"
    description="The requested Lite dashboard page does not exist."
  />
)

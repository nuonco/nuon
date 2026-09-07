import { ComponentDocs } from '../__stories__/ComponentDocs'
import { SubNav } from './SubNav'

export default {
  title: 'lite/molecules/SubNav',
}

const ITEMS = [
  { href: '/', label: 'Connections', end: true },
  { href: '/webhooks', label: 'Webhooks' },
  { href: '/triggers', label: 'Triggers' },
  { href: '/api-tokens', label: 'API tokens' },
  { href: '/service-accounts', label: 'Service accounts' },
  { href: '/oidc', label: 'OIDC federation' },
]

export const Overview = () => (
  <ComponentDocs
    name="SubNav"
    tier="molecule"
    summary="Route-aware navigation between sibling sections."
    use={[
      'Use for related child routes such as organization settings.',
      'Build every href from the owning route context.',
    ]}
    avoid={[
      'Do not use for primary dashboard navigation.',
      'Do not use as a substitute for breadcrumbs.',
    ]}
    rules={[
      'The current route exposes aria-current page.',
      'Overflow remains horizontally scrollable at narrow widths.',
    ]}
  />
)

export const Default = () => (
  <div className="max-w-4xl p-4">
    <SubNav items={ITEMS} label="Settings sections" />
  </div>
)

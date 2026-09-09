import { useState } from 'react'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Dropdown } from '../../atoms/Dropdown'
import { Button } from '../../atoms/Button'
import { OrgSwitcherMenu } from './OrgSwitcherMenu'

export default {
  title: 'lite/organisms/OrgSwitcherMenu',
}

export const Overview = () => (
  <ComponentDocs
    name="OrgSwitcherMenu"
    tier="organism"
    summary="A SwitcherMenu bound to organizations, rendering each as an OrgProfile row."
    use={[
      'Nest inside the user dropdown so account and workspace context stay together.',
      'Use anywhere the active organization needs to change.',
    ]}
    avoid={[
      'Do not add organization management actions such as create or delete.',
      'Do not render organizations that are missing an id or a name.',
    ]}
    rules={[
      'Each row is an OrgProfile, so loading rows share the loaded row anatomy.',
      'Selecting an organization navigates to its root route.',
      'The container fetches on open and clears the search on each open.',
    ]}
    props={[
      {
        name: 'orgs',
        type: 'TOrg[]',
        description: 'Organizations for the current page of results.',
      },
      {
        name: 'currentOrgId',
        type: 'string',
        description: 'Marks the active organization as checked.',
      },
      {
        name: 'search',
        type: 'string',
        description: 'Controlled query for the menu search field.',
      },
      {
        name: 'onSearchChange',
        type: '(value: string) => void',
        description: 'Updates the search query without closing the menu.',
      },
      {
        name: 'onLoadMore',
        type: '() => void',
        description: 'Fetches the next page of organizations without closing the menu.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Shows OrgProfile loading rows.',
      },
      {
        name: 'loadingMore',
        type: 'boolean',
        description: 'Puts the load more row into its pending state.',
      },
      {
        name: 'hasMore',
        type: 'boolean',
        default: 'false',
        description: 'Shows the load more row.',
      },
      {
        name: 'hasError',
        type: 'boolean',
        default: 'false',
        description: 'Replaces the rows with a load failure message.',
      },
    ]}
  />
)

const ORGS = [
  { id: 'orgckabqnhi3x2ixos5sko6js0', name: 'alpha', status: 'active' },
  { id: 'orgd7f2h9k1m3p5r7t9v1x3z5b', name: 'beta', status: 'active' },
  { id: 'orge8g3j0l2n4q6s8u0w2y4a6c', name: 'gamma', status: 'provisioning' },
  { id: 'orgf9h4k1m3p5r7t9v1x3z5b7d', name: 'delta', status: 'active' },
  { id: 'orgg0j5l2n4q6s8u0w2y4a6c8e', name: 'epsilon', status: 'error' },
]

export const Default = () => {
  const [search, setSearch] = useState('')

  return (
    <div className="p-20">
      <Dropdown defaultOpen trigger={<Button>Switch organization</Button>}>
        <OrgSwitcherMenu
          orgs={ORGS.filter((org) => org.name.includes(search))}
          currentOrgId="org_alpha"
          search={search}
          onSearchChange={setSearch}
          onLoadMore={() => {}}
          hasMore
        />
      </Dropdown>
    </div>
  )
}

export const Loading = () => (
  <div className="p-20">
    <Dropdown defaultOpen trigger={<Button>Switch organization</Button>}>
      <OrgSwitcherMenu
        loading
        orgs={[]}
        search=""
        onSearchChange={() => {}}
        onLoadMore={() => {}}
      />
    </Dropdown>
  </div>
)

export const Empty = () => (
  <div className="p-20">
    <Dropdown defaultOpen trigger={<Button>Switch organization</Button>}>
      <OrgSwitcherMenu
        orgs={[]}
        search="nothing"
        onSearchChange={() => {}}
        onLoadMore={() => {}}
      />
    </Dropdown>
  </div>
)

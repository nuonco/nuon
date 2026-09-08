import { useState } from 'react'
import { Dropdown } from '../../atoms/Dropdown'
import { Button } from '../../atoms/Button'
import { OrgSwitcherMenu } from './OrgSwitcherMenu'

export default {
  title: 'lite/organisms/OrgSwitcherMenu',
}

const ORGS = [
  { id: 'org_alpha', name: 'alpha' },
  { id: 'org_beta', name: 'beta' },
  { id: 'org_gamma', name: 'gamma' },
  { id: 'org_delta', name: 'delta' },
  { id: 'org_epsilon', name: 'epsilon' },
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

import { useState } from 'react'
import { Dropdown } from '../atoms/Dropdown'
import { Button } from '../atoms/Button'
import { SwitcherMenu } from './SwitcherMenu'

export default {
  title: 'lite/molecules/SwitcherMenu',
}

const ITEMS = [
  { id: 'br_main', label: 'main', href: '/branches/br_main' },
  { id: 'br_release', label: 'release', href: '/branches/br_release' },
  { id: 'br_preview', label: 'preview', href: '/branches/br_preview' },
  { id: 'br_next', label: 'next', href: '/branches/br_next' },
  { id: 'br_legacy', label: 'legacy', href: '/branches/br_legacy' },
]

export const Default = () => {
  const [search, setSearch] = useState('')

  return (
    <div className="p-20">
      <Dropdown defaultOpen trigger={<Button>Switch branch</Button>}>
        <SwitcherMenu
          items={ITEMS.filter((item) => item.label.includes(search))}
          selectedId="br_main"
          search={search}
          onSearchChange={setSearch}
          onLoadMore={() => {}}
          searchLabel="Search branches"
          searchPlaceholder="Search branches..."
          emptyTitle="No branches found"
          errorTitle="Branches failed to load"
          hasMore
        />
      </Dropdown>
    </div>
  )
}

export const Loading = () => (
  <div className="p-20">
    <Dropdown defaultOpen trigger={<Button>Switch organization</Button>}>
      <SwitcherMenu
        items={[]}
        search=""
        onSearchChange={() => {}}
        onLoadMore={() => {}}
        searchLabel="Search organizations"
        searchPlaceholder="Search organizations..."
        emptyTitle="No organizations found"
        errorTitle="Organizations failed to load"
        isLoading
      />
    </Dropdown>
  </div>
)

import { useState } from 'react'
import { BranchSwitcher } from './BranchSwitcher'

export default {
  title: 'lite/organisms/BranchSwitcher',
}

const BRANCHES = [
  { id: 'br_main', name: 'main' },
  { id: 'br_release', name: 'release' },
  { id: 'br_preview', name: 'preview' },
  { id: 'br_next', name: 'next' },
  { id: 'br_legacy', name: 'legacy' },
]

export const Default = () => {
  const [search, setSearch] = useState('')

  return (
    <div className="flex justify-end p-20">
      <BranchSwitcher
        branches={BRANCHES.filter((branch) => branch.name.includes(search))}
        currentBranch={BRANCHES[0]}
        search={search}
        onSearchChange={setSearch}
        onLoadMore={() => {}}
        getBranchHref={(branchId) => `/branches/${branchId}`}
        hasMore
      />
    </div>
  )
}

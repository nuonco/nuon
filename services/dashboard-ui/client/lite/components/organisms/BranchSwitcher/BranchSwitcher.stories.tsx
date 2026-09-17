import { useState } from 'react'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { BranchSwitcher } from './BranchSwitcher'

export default {
  title: 'lite/organisms/BranchSwitcher',
}

export const Overview = () => (
  <ComponentDocs
    name="BranchSwitcher"
    tier="organism"
    summary="A branch-name trigger opening a SwitcherMenu of the app's branches."
    use={[
      'Place beside the app branch subnav so the active branch is always visible.',
      'Use anywhere the reader needs to move between branches of one app.',
    ]}
    avoid={[
      'Do not use it to create, configure, or delete a branch.',
      'Do not render branches that are missing an id or a name.',
    ]}
    rules={[
      'The trigger shows the current branch name in mono, or Select branch when none resolves.',
      'Switching branches preserves the current subsection through getBranchHref.',
      'The container fetches on open and clears the search on each open.',
      'Search, paging and error state are passed through to SwitcherMenu.',
    ]}
    props={[
      {
        name: 'branches',
        type: 'TAppBranch[]',
        description: 'Branches for the current page of results.',
      },
      {
        name: 'currentBranch',
        type: 'TAppBranch',
        description: 'Branch named in the trigger and checked in the menu.',
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
        description: 'Fetches the next page of branches without closing the menu.',
      },
      {
        name: 'getBranchHref',
        type: '(branchId: string) => string',
        description:
          'Builds each row destination, carrying the active subsection across.',
      },
      {
        name: 'onOpenChange',
        type: '(open: boolean) => void',
        description:
          'Reports the dropdown open state so the container can gate fetching.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Loads the trigger and shows menu loading rows.',
      },
      {
        name: 'loadingMore',
        type: 'boolean',
        description: 'Puts the load more row into its pending state.',
      },
      {
        name: 'hasMore',
        type: 'boolean',
        description: 'Shows the load more row.',
      },
      {
        name: 'hasError',
        type: 'boolean',
        description: 'Replaces the rows with a load failure message.',
      },
    ]}
  />
)

const BRANCHES = [
  'main',
  'release',
  'preview',
  'next',
  'feature/long-running-migration',
].map((name, index) => ({
  id: `brncq7fplr1up5atx5zpxotbab${index}`,
  name,
  configs: [
    {
      config_number: 14 - index,
      connected_github_vcs_config: { directory: 'services/payments' },
    },
  ],
}))

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

export const Loading = () => (
  <div className="flex justify-end p-20">
    <BranchSwitcher
      loading
      branches={[]}
      currentBranch={BRANCHES[0]}
      search=""
      onSearchChange={() => {}}
      onLoadMore={() => {}}
      getBranchHref={(branchId) => `/branches/${branchId}`}
    />
  </div>
)

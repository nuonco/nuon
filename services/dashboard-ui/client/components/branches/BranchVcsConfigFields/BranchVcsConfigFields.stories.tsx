export default {
  title: 'Branches/BranchVcsConfigFields',
}

import { useState } from 'react'
import { BranchVcsConfigFields } from './BranchVcsConfigFields'

const noop = () => {}

const connections = [
  { id: 'vcs-1', github_account_name: 'acme-co', github_install_id: '12345' },
] as any

const repos = [
  { full_name: 'acme-co/api', private: true },
  { full_name: 'acme-co/web', private: true },
  { full_name: 'acme-co/public-docs', private: false },
] as any

const branches = [{ name: 'main' }, { name: 'develop' }, { name: 'release/1.x' }] as any

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div className="flex flex-col gap-4 max-w-md">{children}</div>
)

export const Default = () => {
  const [repo, setRepo] = useState<any>(repos[0])
  const [branch, setBranch] = useState('main')
  const [directory, setDirectory] = useState('.')
  const [pathFilter, setPathFilter] = useState('')
  return (
    <Wrap>
      <BranchVcsConfigFields
        vcsConnections={connections}
        repos={repos}
        branches={branches}
        loadingRepos={false}
        loadingBranches={false}
        reposError={null}
        branchesError={null}
        selectedVcsConnectionId="vcs-1"
        onVcsConnectionChange={noop}
        selectedRepo={repo}
        onRepoChange={setRepo}
        selectedBranch={branch}
        onBranchChange={setBranch}
        directory={directory}
        onDirectoryChange={setDirectory}
        pathFilter={pathFilter}
        onPathFilterChange={setPathFilter}
        isSubmitting={false}
      />
    </Wrap>
  )
}

export const LoadingRepos = () => (
  <Wrap>
    <BranchVcsConfigFields
      vcsConnections={connections}
      repos={[]}
      branches={[]}
      loadingRepos
      loadingBranches={false}
      reposError={null}
      branchesError={null}
      selectedVcsConnectionId="vcs-1"
      onVcsConnectionChange={noop}
      selectedRepo={null}
      onRepoChange={noop}
      selectedBranch=""
      onBranchChange={noop}
      directory="."
      onDirectoryChange={noop}
      pathFilter=""
      onPathFilterChange={noop}
      isSubmitting={false}
    />
  </Wrap>
)

export const NoConnections = () => (
  <Wrap>
    <BranchVcsConfigFields
      vcsConnections={[]}
      repos={[]}
      branches={[]}
      loadingRepos={false}
      loadingBranches={false}
      reposError={null}
      branchesError={null}
      selectedVcsConnectionId=""
      onVcsConnectionChange={noop}
      selectedRepo={null}
      onRepoChange={noop}
      selectedBranch=""
      onBranchChange={noop}
      directory="."
      onDirectoryChange={noop}
      pathFilter=""
      onPathFilterChange={noop}
      isSubmitting={false}
    />
  </Wrap>
)

export const ManualBranchEntry = () => {
  const [branch, setBranch] = useState('')
  return (
    <Wrap>
      <BranchVcsConfigFields
        vcsConnections={connections}
        repos={repos}
        branches={[]}
        loadingRepos={false}
        loadingBranches={false}
        reposError={null}
        branchesError={null}
        selectedVcsConnectionId="vcs-1"
        onVcsConnectionChange={noop}
        selectedRepo={repos[0]}
        onRepoChange={noop}
        selectedBranch={branch}
        onBranchChange={setBranch}
        directory="."
        onDirectoryChange={noop}
        pathFilter=""
        onPathFilterChange={noop}
        isSubmitting={false}
      />
    </Wrap>
  )
}

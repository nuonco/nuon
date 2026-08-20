export default {
  title: 'Installs/InstallBranchFilter',
}

import { InstallBranchFilter } from './InstallBranchFilter'

const branchNames = ['main', 'develop', 'staging', 'release']

export const Default = () => (
  <InstallBranchFilter
    queryKey={['org-branch-names', 'story']}
    queryFn={() => Promise.resolve(branchNames)}
  />
)

export const NoBranches = () => (
  <InstallBranchFilter
    queryKey={['org-branch-names', 'story-empty']}
    queryFn={() => Promise.resolve([])}
  />
)

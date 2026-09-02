import { BranchVcsBadges } from './BranchVcsBadges'

export default {
  title: 'Branches/BranchVcsBadges',
}

export const Default = () => (
  <span className="flex items-center gap-2 flex-wrap">
    <BranchVcsBadges
      repo="nuonco/example-app-configs"
      branch="nh/test-app-branches"
    />
  </span>
)

export const RepoOnly = () => (
  <span className="flex items-center gap-2 flex-wrap">
    <BranchVcsBadges repo="nuonco/example-app-configs" />
  </span>
)

export const BranchOnly = () => (
  <span className="flex items-center gap-2 flex-wrap">
    <BranchVcsBadges branch="main" />
  </span>
)

export const Empty = () => (
  <span className="flex items-center gap-2 flex-wrap">
    <BranchVcsBadges />
  </span>
)

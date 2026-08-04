import type { TAppBranch } from '@/types'
import { AppBranchSwitcher } from './AppBranchSwitcher'

export default {
  title: 'Branches/AppBranchSwitcher',
}

const mockBranches: TAppBranch[] = [
  { id: 'brnch_1', name: 'test', app_id: 'app_1', org_id: 'org_1' },
  { id: 'brnch_2', name: 'main', app_id: 'app_1', org_id: 'org_1' },
  { id: 'brnch_3', name: 'nh/test-app-branches', app_id: 'app_1', org_id: 'org_1' },
  { id: 'brnch_4', name: 'staging', app_id: 'app_1', org_id: 'org_1' },
]

const currentBranch = mockBranches[0]

export const Default = () => (
  <AppBranchSwitcher
    branches={mockBranches}
    currentBranch={currentBranch}
    orgId="org_1"
    appId="app_1"
    isLoading={false}
  />
)

export const Loading = () => (
  <AppBranchSwitcher
    branches={[]}
    currentBranch={currentBranch}
    orgId="org_1"
    appId="app_1"
    isLoading
  />
)

export const WithSearch = () => (
  <AppBranchSwitcher
    branches={Array.from({ length: 12 }).map((_, i) => ({
      id: `brnch_${i}`,
      name: i === 0 ? 'test' : `feature/branch-${i}`,
      app_id: 'app_1',
      org_id: 'org_1',
    }))}
    currentBranch={currentBranch}
    orgId="org_1"
    appId="app_1"
    isLoading={false}
  />
)

export const SingleBranch = () => (
  <AppBranchSwitcher
    branches={[currentBranch]}
    currentBranch={currentBranch}
    orgId="org_1"
    appId="app_1"
    isLoading={false}
  />
)

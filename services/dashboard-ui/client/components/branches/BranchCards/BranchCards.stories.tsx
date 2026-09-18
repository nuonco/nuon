export default {
  title: 'Branches/BranchCards',
}

import { BranchCards } from './BranchCards'
import type { TBranchCardData } from './BranchCard'

const mockCards: TBranchCardData[] = [
  {
    branchId: 'br-001',
    name: 'main',
    href: '/org-1/apps/app-1/branches/br-001',
    repo: 'acme/platform-configs',
    repoBranch: 'main',
    latestRun: {
      href: '/org-1/apps/app-1/branches/br-001/runs/run-1',
      status: 'success',
      commitMessage: 'bump api image to v1.42.0',
      sha: '85d067ecafe1234',
      author: 'Example Developer',
      createdAt: '2026-07-30T10:30:00Z',
    },
    planGroups: [
      { name: 'staging', installs: 2, hasSelector: false },
      { name: 'customers', installs: 4, hasSelector: false },
      { name: 'enterprise', installs: 1, hasSelector: false },
    ],
  },
  {
    branchId: 'br-002',
    name: 'staging',
    href: '/org-1/apps/app-1/branches/br-002',
    repo: 'acme/platform-configs',
    repoBranch: 'staging',
    latestRun: {
      href: '/org-1/apps/app-1/branches/br-002/runs/run-2',
      status: 'running',
      commitMessage: 'add new worker component',
      sha: '6987a43568abc82',
      author: 'Example Developer',
      createdAt: '2026-08-01T14:00:00Z',
      awaitingApproval: true,
    },
    planGroups: [{ name: 'customers', installs: 0, hasSelector: true }],
  },
  {
    branchId: 'br-003',
    name: 'feature/new-deploy',
    href: '/org-1/apps/app-1/branches/br-003',
    planGroups: [],
  },
]

export const Default = () => <BranchCards cards={mockCards} />

export const Loading = () => <BranchCards cards={[]} isLoading />

export const Empty = () => <BranchCards cards={[]} />

export const WithPagination = () => (
  <BranchCards
    cards={mockCards}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

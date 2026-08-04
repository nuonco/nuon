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
    managedBy: 'config',
    repo: 'acme/platform-configs',
    repoBranch: 'main',
    latestRun: {
      href: '/org-1/apps/app-1/branches/br-001/runs/run-1',
      status: 'success',
      commitMessage: 'bump api image to v1.42.0',
      createdAt: '2026-07-30T10:30:00Z',
    },
    planSummary: { groups: 2, installs: 7, hasSelector: false },
  },
  {
    branchId: 'br-002',
    name: 'staging',
    href: '/org-1/apps/app-1/branches/br-002',
    managedBy: 'manually',
    repo: 'acme/platform-configs',
    repoBranch: 'staging',
    latestRun: {
      href: '/org-1/apps/app-1/branches/br-002/runs/run-2',
      status: 'running',
      commitMessage: 'add new worker component',
      createdAt: '2026-08-01T14:00:00Z',
      awaitingApproval: true,
    },
    planSummary: { groups: 1, installs: 0, hasSelector: true },
  },
  {
    branchId: 'br-003',
    name: 'feature/new-deploy',
    href: '/org-1/apps/app-1/branches/br-003',
    managedBy: null,
    planSummary: { groups: 0, installs: 0, hasSelector: false },
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

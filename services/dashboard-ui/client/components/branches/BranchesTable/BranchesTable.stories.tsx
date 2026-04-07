export default {
  title: 'Branches/BranchesTable',
}

import { BranchesTable } from './BranchesTable'

const mockData = [
  {
    branchId: 'br-001',
    branchName: 'production',
    workflowCount: 3,
    createdAt: '2024-01-15T10:30:00Z',
    href: '/org-1/apps/app-1/branches/br-001',
  },
  {
    branchId: 'br-002',
    branchName: 'staging',
    workflowCount: 1,
    createdAt: '2024-02-20T14:00:00Z',
    href: '/org-1/apps/app-1/branches/br-002',
  },
  {
    branchId: 'br-003',
    branchName: 'feature/new-deploy',
    workflowCount: 0,
    createdAt: '2024-03-01T09:00:00Z',
    href: '/org-1/apps/app-1/branches/br-003',
  },
]

export const Default = () => (
  <BranchesTable data={mockData} isLoading={false} />
)

export const Loading = () => (
  <BranchesTable data={[]} isLoading={true} />
)

export const Empty = () => (
  <BranchesTable data={[]} isLoading={false} />
)

export const SingleWorkflow = () => (
  <BranchesTable
    data={[
      {
        branchId: 'br-001',
        branchName: 'main',
        workflowCount: 1,
        createdAt: '2024-01-15T10:30:00Z',
        href: '/org-1/apps/app-1/branches/br-001',
      },
    ]}
    isLoading={false}
  />
)

export const WithPagination = () => (
  <BranchesTable
    data={mockData}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

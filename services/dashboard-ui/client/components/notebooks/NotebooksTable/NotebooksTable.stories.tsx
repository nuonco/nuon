export default {
  title: 'Notebooks/NotebooksTable',
}

import {
  NotebooksTable,
  NotebooksTableSkeleton,
  type TNotebookRow,
} from './NotebooksTable'

const mockRows: TNotebookRow[] = [
  {
    id: 'nbk-1',
    name: 'Debug pods',
    description: (
      <span className="text-sm text-gray-500">
        Commands for debugging pod issues in production.
      </span>
    ),
    cellCount: 4,
    lastRun: <span className="text-sm text-gray-500">2 minutes ago</span>,
    lastUpdated: <span className="text-sm text-gray-500">5 minutes ago</span>,
    href: '/org-1/installs/install-1/notebooks/nbk-1',
  },
  {
    id: 'nbk-2',
    name: 'Rotate secrets',
    description: null,
    cellCount: 1,
    lastRun: null,
    lastUpdated: <span className="text-sm text-gray-500">1 hour ago</span>,
    href: '/org-1/installs/install-1/notebooks/nbk-2',
  },
  {
    id: 'nbk-3',
    name: 'Health checks',
    description: (
      <span className="text-sm text-gray-500">
        Run health checks across all services.
      </span>
    ),
    cellCount: 7,
    lastRun: <span className="text-sm text-gray-500">3 days ago</span>,
    lastUpdated: <span className="text-sm text-gray-500">3 days ago</span>,
    href: '/org-1/installs/install-1/notebooks/nbk-3',
  },
]

const pagination = { hasNext: false, offset: 0, limit: 20 }

export const Default = () => (
  <NotebooksTable data={mockRows} isLoading={false} pagination={pagination} />
)

export const Empty = () => (
  <NotebooksTable data={[]} isLoading={false} pagination={pagination} />
)

export const Loading = () => <NotebooksTableSkeleton />

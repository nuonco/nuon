export default {
  title: 'Components/ComponentsTable',
}

import { ComponentsTable, ComponentsTableSkeleton, type TComponentRow } from './ComponentsTable'

const mockRows: TComponentRow[] = Array.from({ length: 3 }, (_, i) => ({
  buildStatus: <span>active</span>,
  componentId: `comp-${i + 1}`,
  componentName: `Component ${i + 1}`,
  componentType: <span>terraform_module</span>,
  href: `/org-1/apps/app-1/components/comp-${i + 1}`,
  dependencies: <span>-</span>,
  labels: i === 0 ? <span className="text-xs font-mono">tier: infrastructure</span> : null,
}))

export const Default = () => (
  <ComponentsTable
    data={mockRows}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <ComponentsTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <ComponentsTableSkeleton />

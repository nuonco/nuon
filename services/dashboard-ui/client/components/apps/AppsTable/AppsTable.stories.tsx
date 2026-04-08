export default {
  title: 'Apps/AppsTable',
}

import { AppsTable, AppsTableSkeleton, type TAppRow } from './AppsTable'

const mockRows: TAppRow[] = Array.from({ length: 3 }, (_, i) => ({
  actionHref: `/org-1/apps/app-${i + 1}`,
  appId: `app-${i + 1}`,
  configVersion: i + 1,
  defaultBranch: 'main',
  name: `My App ${i + 1}`,
  nameHref: `/org-1/apps/app-${i + 1}`,
  platform: 'aws',
  sandboxHref: `https://github.com/nuonco/app-${i + 1}`,
  sandboxName: `nuonco/app-${i + 1}`,
}))

export const Default = () => (
  <AppsTable
    data={mockRows}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <AppsTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <AppsTableSkeleton />

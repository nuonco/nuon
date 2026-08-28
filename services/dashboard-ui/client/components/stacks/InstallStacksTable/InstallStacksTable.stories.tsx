import { InstallStacksTable, type TStackVersion } from './InstallStacksTable'

export default {
  title: 'Stacks/InstallStacksTable',
}

const mockVersions = Array.from({ length: 3 }, (_, i) => ({
  id: `stkv-${i + 1}`,
  app_config_id: `cfg-${i + 1}`,
  created_at: new Date(Date.now() - i * 86400000).toISOString(),
  updated_at: new Date(Date.now() - i * 43200000).toISOString(),
  composite_status: {
    status: i === 0 ? 'active' : 'expired',
    history: [{ status: 'pending', created_at_ts: Date.now() - 7200000 }],
  },
  runs: Array.from({ length: i + 1 }, (_, r) => ({
    id: `run-${i}-${r}`,
    created_at: new Date(Date.now() - r * 3600000).toISOString(),
    run_type: r === 0 ? 'workflow-run' : 'out-of-band',
    data_contents: { vpc_id: 'vpc-abc123' },
  })),
  quick_link_url: 'https://console.aws.amazon.com/cloudformation/home',
  template_url: 'https://s3.amazonaws.com/nuon-stacks/template.json',
})) as unknown as TStackVersion[]

export const Default = () => (
  <InstallStacksTable versions={mockVersions} orgId="org-1" appId="app-1" />
)

export const Empty = () => (
  <InstallStacksTable versions={[]} orgId="org-1" appId="app-1" />
)

export const Loading = () => (
  <InstallStacksTable versions={[]} orgId="org-1" appId="app-1" isLoading />
)

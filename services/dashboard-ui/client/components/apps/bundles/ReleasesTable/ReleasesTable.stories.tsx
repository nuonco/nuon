import { ReleasesTable } from './ReleasesTable'

export default {
  title: 'Apps/Bundles/ReleasesTable',
}

const releases = [
  {
    id: 'apr123456789',
    app_config_id: 'abc123456789',
    created_at: '2026-08-31T12:00:00Z',
    semantic_digest: 'sha256:87c21aa1111111111111111111111111111111111111111111111111111111',
    status: 'ready',
  },
]

export const Default = () => (
  <ReleasesTable
    appId="app123"
    data={releases}
    orgId="org123"
    pagination={{ limit: 20, offset: 0 }}
  />
)

export const Empty = () => (
  <ReleasesTable
    appId="app123"
    data={[]}
    orgId="org123"
    pagination={{ limit: 20, offset: 0 }}
  />
)

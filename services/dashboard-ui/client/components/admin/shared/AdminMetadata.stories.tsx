export default {
  title: 'Admin/AdminMetadata',
}

import { AdminInfoCard, AdminMetadataPanel } from './AdminMetadata'

export const InfoCard = () => (
  <AdminInfoCard
    title="Install ID"
    value="inst_abc123def456"
    copyable
  />
)

export const InfoCardLoading = () => (
  <AdminInfoCard
    title="Org ID"
    value={null}
    loading
  />
)

export const InfoCardEmpty = () => (
  <AdminInfoCard
    title="Temporal Workflow ID"
    value={null}
  />
)

export const MetadataPanel = () => (
  <AdminMetadataPanel>
    <AdminInfoCard title="Install ID" value="inst_abc123def456" copyable />
    <AdminInfoCard title="App ID" value="app_xyz789" copyable />
    <AdminInfoCard title="Org ID" value="org_qrs456" copyable />
    <AdminInfoCard title="Status" value="active" />
  </AdminMetadataPanel>
)

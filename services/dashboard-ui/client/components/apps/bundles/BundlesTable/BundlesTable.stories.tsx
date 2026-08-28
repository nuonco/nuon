import type { TCustomerManagedBundle } from '@/types'
import { BundlesTable, BundlesTableSkeleton } from './BundlesTable'

export default {
  title: 'Apps/Bundles/BundlesTable',
}

const mockBundles: TCustomerManagedBundle[] = [
  {
    id: 'agbrzvk1l0ye0rtb0y4nmhs1lv',
    app_id: 'appke1qx2unrp2a0onzb2jy7uv',
    app_config_id: 'app608mmtt1p456k8f65znhl43',
    status: 'active',
    status_description: 'bundle published and verified',
    target_platform: 'linux/amd64',
    size: 390120275,
    created_at: '2026-08-01T10:00:00Z',
    transport_checksum:
      '395dfcd8e26c7ff5d80297715274ac18366e51d631c6de4bf7f30bafb618767e',
  },
  {
    id: 'agbg1fbl9enzmsttpiidpplx0k',
    app_id: 'appke1qx2unrp2a0onzb2jy7uv',
    status: 'publishing',
    status_description: 'uploading bundle archive',
    target_platform: 'linux/amd64',
    created_at: '2026-08-02T11:30:00Z',
  },
  {
    id: 'agbv37ghrbw83dpf7krid4vj0x',
    app_id: 'appke1qx2unrp2a0onzb2jy7uv',
    status: 'error',
    status_description: 'bundle publish failed',
    target_platform: 'linux/amd64',
    created_at: '2026-08-03T09:15:00Z',
  },
]

export const Default = () => (
  <BundlesTable
    data={mockBundles}
    isLoading={false}
    orgId="org-mock-001"
    appId="appke1qx2unrp2a0onzb2jy7uv"
  />
)

export const Empty = () => <BundlesTable data={[]} isLoading={false} />

export const Loading = () => <BundlesTableSkeleton />

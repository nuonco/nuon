export default {
  title: 'Policies/PoliciesTable',
}

import { PoliciesTable } from './PoliciesTable'
import type { TAppPolicyConfig } from '@/types'

const mockPolicies: TAppPolicyConfig[] = [
  {
    id: 'pol-1',
    name: 'no-privileged-containers',
    type: 'admission_control',
    engine: 'kyverno',
    components: ['*'],
    contents: btoa('metadata:\n  name: no-privileged-containers'),
    created_at: '2024-01-01T00:00:00Z',
  } as unknown as TAppPolicyConfig,
  {
    id: 'pol-2',
    name: 'resource-limits',
    type: 'admission_control',
    engine: 'opa',
    components: ['api', 'worker'],
    contents: btoa('package resource_limits'),
    created_at: '2024-01-02T00:00:00Z',
  } as unknown as TAppPolicyConfig,
]

export const Default = () => (
  <PoliciesTable policies={mockPolicies} orgId="org-1" appId="app-1" />
)

export const Empty = () => (
  <PoliciesTable policies={[]} orgId="org-1" appId="app-1" />
)

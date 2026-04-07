export default {
  title: 'Stacks/StackOutputs',
}

import { StackOutputs } from './StackOutputs'
import type { TInstallStackVersionRun } from '@/types'

const mockRuns: TInstallStackVersionRun[] = [
  {
    id: 'run-1',
    created_at: new Date(Date.now() - 300000).toISOString(),
    data: {
      vpc_id: 'vpc-0abc123',
      subnet_ids: 'subnet-1,subnet-2',
      cluster_endpoint: 'https://k8s.example.com',
    },
  } as TInstallStackVersionRun,
  {
    id: 'run-2',
    created_at: new Date(Date.now() - 900000).toISOString(),
    data: {
      vpc_id: 'vpc-old123',
    },
  } as TInstallStackVersionRun,
]

export const Default = () => <StackOutputs runs={mockRuns} />

export const Single = () => <StackOutputs runs={[mockRuns[0]]} />

export const Empty = () => <StackOutputs runs={[]} />

export default {
  title: 'Stacks/InstallStackVersions',
}

import { Button } from '@/components/common/Button'
import {
  InstallStackVersions,
  type TInstallStackVersion,
} from './InstallStackVersions'

const mockVersions = [
  {
    id: 'stkv-1',
    app_config_id: 'cfg-3',
    created_at: '2026-09-20T18:04:00Z',
    composite_status: { status: 'active' },
    runs: [{ id: 'run-1' }],
  },
  {
    id: 'stkv-2',
    app_config_id: 'cfg-2',
    created_at: '2026-09-18T12:00:00Z',
    composite_status: { status: 'active' },
    runs: [{ id: 'run-2' }, { id: 'run-3' }],
  },
  {
    id: 'stkv-3',
    app_config_id: 'cfg-1',
    created_at: '2026-09-10T08:00:00Z',
    composite_status: { status: 'expired' },
    runs: [],
  },
] as unknown as TInstallStackVersion[]

const latestAction = (
  <Button variant="secondary" size="sm">
    Reprovision stack
  </Button>
)

export const Default = () => (
  <InstallStackVersions versions={mockVersions} latestAction={latestAction} />
)

export const Empty = () => <InstallStackVersions versions={[]} />

export const Loading = () => <InstallStackVersions versions={[]} loading />

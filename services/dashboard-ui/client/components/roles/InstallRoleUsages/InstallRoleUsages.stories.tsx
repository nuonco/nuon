export default {
  title: 'Roles/InstallRoleUsages',
}

import { ModalStory } from '@/components/__stories__/helpers'
import {
  InstallRoleUsagesModal,
  InstallRoleUsagesTrigger,
} from './InstallRoleUsages'

const mockUsages = [
  {
    id: 'iru1',
    role_name: 'inl123-alb-deploy',
    role_source: 'entity',
    runner_job: {
      id: 'job1',
      type: 'helm-chart-deploy',
      owner_type: 'install_deploys',
      owner_id: 'dpl1',
      created_at: '2026-04-20T10:00:00Z',
      metadata: { component_name: 'application_load_balancer' },
    },
    workflow: { id: 'inw1', name: 'Provisioned install', type: 'provision' },
    workflow_step_id: 'iws1',
  },
  {
    id: 'iru2',
    role_name: 'inl123-alb-deploy',
    role_source: 'default',
    runner_job: {
      id: 'job2',
      type: 'sandbox-terraform',
      owner_type: 'install_sandbox_runs',
      owner_id: 'sbr1',
      created_at: '2026-04-19T14:30:00Z',
      metadata: { sandbox_run_type: 'provision' },
    },
    workflow: { id: 'inw2', name: 'Reprovisioned sandbox', type: 'reprovision' },
    workflow_step_id: 'iws2',
  },
] as any

export const Default = () => (
  <ModalStory>
    <InstallRoleUsagesModal
      orgId="org1"
      installId="inl1"
      usages={mockUsages}
      isLoading={false}
      error={null}
      roleDisplayName="alb deploy"
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <InstallRoleUsagesModal
      orgId="org1"
      installId="inl1"
      usages={undefined}
      isLoading={true}
      error={null}
    />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory>
    <InstallRoleUsagesModal
      orgId="org1"
      installId="inl1"
      usages={[]}
      isLoading={false}
      error={null}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <InstallRoleUsagesModal
      orgId="org1"
      installId="inl1"
      usages={undefined}
      isLoading={false}
      error={{ error: 'Unable to load role usage.' } as any}
    />
  </ModalStory>
)

export const Trigger = () => (
  <InstallRoleUsagesTrigger onOpenModal={() => alert('open')} />
)

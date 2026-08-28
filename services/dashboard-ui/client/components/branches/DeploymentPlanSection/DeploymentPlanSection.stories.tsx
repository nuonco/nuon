import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { DeploymentPlanSection } from './DeploymentPlanSection'

export default {
  title: 'Branches/DeploymentPlanSection',
}

const installsById: Record<string, any> = {
  'inst-1': {
    id: 'inst-1',
    name: 'acme-prod',
    labels: { env: 'production', tier: 'primary' },
    status_v2: { status: 'installed' },
  },
  'inst-2': {
    id: 'inst-2',
    name: 'acme-eu',
    labels: { env: 'production', tier: 'secondary' },
    status_v2: { status: 'installed' },
  },
  'inst-3': {
    id: 'inst-3',
    name: 'acme-staging',
    labels: { env: 'staging' },
    status_v2: { status: 'deploying' },
  },
  'inst-4': {
    id: 'inst-4',
    name: 'demo-env',
    labels: {},
    status_v2: { status: 'installed' },
  },
}

const CreateAction = () => (
  <Button variant="secondary">
    <Icon variant="SlidersHorizontalIcon" size={16} />
    Create deployment plan
  </Button>
)

const EditAction = () => (
  <Button variant="ghost">
    <Icon variant="SlidersHorizontalIcon" size={16} />
    Edit plan
  </Button>
)

export const Empty = () => (
  <DeploymentPlanSection
    installsById={{}}
    orgId="org-1"
    createAction={<CreateAction />}
  />
)

export const SingleGroup = () => (
  <DeploymentPlanSection
    config={
      {
        config_number: 1,
        install_groups: [
          {
            id: 'group-1',
            name: 'All installs',
            order: 0,
            max_parallel: 1,
            install_ids: ['inst-1', 'inst-4'],
          },
        ],
      } as any
    }
    installsById={installsById}
    orgId="org-1"
    createAction={<CreateAction />}
    editAction={<EditAction />}
  />
)

export const MultipleGroups = () => (
  <DeploymentPlanSection
    config={
      {
        config_number: 3,
        install_groups: [
          {
            id: 'group-canary',
            name: 'Canary',
            order: 0,
            max_parallel: 1,
            install_ids: ['inst-4'],
          },
          {
            id: 'group-staging',
            name: 'Staging',
            order: 1,
            max_parallel: 1,
            install_ids: ['inst-3'],
          },
          {
            id: 'group-prod',
            name: 'Production',
            order: 2,
            max_parallel: 2,
            install_ids: ['inst-1', 'inst-2'],
          },
        ],
      } as any
    }
    installsById={installsById}
    orgId="org-1"
    createAction={<CreateAction />}
    editAction={<EditAction />}
  />
)

export const LabelBased = () => (
  <DeploymentPlanSection
    config={
      {
        config_number: 2,
        install_groups: [
          {
            id: 'group-staging',
            name: 'Staging',
            order: 0,
            max_parallel: 2,
            label_selector: { match_labels: { env: 'staging' } },
          },
          {
            id: 'group-prod',
            name: 'Production',
            order: 1,
            max_parallel: 1,
            label_selector: { match_labels: { env: 'production' } },
          },
        ],
      } as any
    }
    installsById={installsById}
    orgId="org-1"
    createAction={<CreateAction />}
    editAction={<EditAction />}
  />
)

export default {
  title: 'Branches/WorkflowStepDetail/PlanGroupStep',
}

import type { ComponentProps } from 'react'
import { Button } from '@/components/common/Button'
import { StepCardStory } from '@/components/__stories__/helpers'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { PlanGroupStep, type PlanInstallDiff } from './PlanGroupStep'

const ec2ChangeSection: DiffSectionData = {
  name: 'Components',
  sectionKey: 'components',
  grouped: true,
  additions: 0,
  removals: 0,
  changed: 1,
  entities: [
    {
      name: 'ec2',
      op: 'change',
      componentType: 'terraform_module',
      fields: [
        {
          key: 'instance_type',
          op: 'change',
          diff: '- instance_type: t3.small\n+ instance_type: t3.medium',
        },
      ],
    },
  ],
  fields: [],
}

const mixedSection: DiffSectionData = {
  name: 'Components',
  sectionKey: 'components',
  grouped: true,
  additions: 1,
  removals: 1,
  changed: 1,
  entities: [
    {
      name: 'redis',
      op: 'add',
      componentType: 'helm_chart',
      fields: [{ key: 'image', op: 'add', diff: '+ image: redis:7' }],
    },
    {
      name: 'ec2',
      op: 'change',
      componentType: 'terraform_module',
      fields: [
        {
          key: 'instance_type',
          op: 'change',
          diff: '- instance_type: t3.small\n+ instance_type: t3.medium',
        },
      ],
    },
    {
      name: 'legacy_worker',
      op: 'remove',
      componentType: 'docker_build',
      fields: [{ key: 'image', op: 'remove', diff: '- image: legacy:1' }],
    },
  ],
  fields: [],
}

const uatInstalls: PlanInstallDiff[] = [
  {
    installId: 'inlex6ib6i2yvkqpjt643h63fb',
    installName: 'httpbin_dev',
    sections: [],
    summary: { added: 0, removed: 0, changed: 0 },
  },
  {
    installId: 'inlrlganokl7ffhxr4jtje08sz',
    installName: 'test',
    sections: [],
    summary: { added: 0, removed: 0, changed: 0 },
  },
  {
    installId: 'inlq8tqxoz0jtdigksj3nnebhc',
    installName: 'testtest',
    sections: [ec2ChangeSection],
    summary: { added: 0, removed: 0, changed: 1 },
  },
]

const labeledInstalls: PlanInstallDiff[] = [
  {
    installId: 'inlacmeprod',
    installName: 'acme-prod',
    installLabels: { tier: 'prod', region: 'us-east-1' },
    sections: [mixedSection],
    summary: { added: 1, removed: 1, changed: 1 },
  },
  {
    installId: 'inlacmestg',
    installName: 'acme-staging',
    installLabels: { tier: 'staging', region: 'us-west-2' },
    sections: [],
    summary: { added: 0, removed: 0, changed: 0 },
  },
]

const overflowingInstalls: PlanInstallDiff[] = [
  {
    installId: 'inlwsworkspace01kzree4e4e',
    installName: 'ws-workspace_01kzree4e4ejrsmaj9vbs4j0mg-sandbox-fd-aug-11',
    installLabels: {
      'auto-deploy': 'true',
      canary: 'true',
      install_id: 'byoci_266gevfxpr9m7a7mpawv8vd01q',
      workspace_id: 'workspace_01kzree4e4ejrsmaj9vbs4j0mg',
    },
    sections: [],
    summary: { added: 0, removed: 0, changed: 0 },
  },
  {
    installId: 'inlwsworkspace09ffb2c1aa22',
    installName: 'ws-workspace_09ffb2c1aa22ppld5xkq7t1nz-sandbox-gh-sep-02',
    installLabels: {
      'auto-deploy': 'false',
      canary: 'false',
      environment: 'preproduction-eu-central',
      install_id: 'byoci_744mnbvcx8qw3e5r6t7y8u9i0p',
      workspace_id: 'workspace_09ffb2c1aa22ppld5xkq7t1nz',
    },
    sections: [mixedSection],
    summary: { added: 1, removed: 1, changed: 1 },
  },
]

const actions = (
  <>
    <Button variant="danger">Skip</Button>
    <Button variant="primary">Approve</Button>
  </>
)

const StepInCard = (props: ComponentProps<typeof PlanGroupStep>) => (
  <StepCardStory name="plan install group" status="awaiting-approval">
    <PlanGroupStep {...props} />
  </StepCardStory>
)

export const AwaitingApproval = () => (
  <StepInCard
    installs={uatInstalls}
    groupName="uat"
    hasResponse={false}
    showApproveBar
    isInProgress={false}
    actions={actions}
  />
)

export const SingleInstall = () => (
  <StepInCard
    installs={[uatInstalls[2]]}
    groupName="production"
    hasResponse={false}
    showApproveBar
    isInProgress={false}
    actions={actions}
  />
)

export const WithLabels = () => (
  <StepInCard
    installs={labeledInstalls}
    groupName="production"
    hasResponse={false}
    showApproveBar
    isInProgress={false}
    actions={actions}
  />
)

export const OverflowingLabels = () => (
  <StepInCard
    installs={overflowingInstalls}
    groupName="canary"
    hasResponse
    responseType="approve"
    showApproveBar={false}
    isInProgress={false}
  />
)

export const Loading = () => (
  <StepInCard
    installs={uatInstalls.map((i) => ({
      ...i,
      sections: [],
      summary: null,
      isLoading: true,
    }))}
    groupName="uat"
    hasResponse={false}
    showApproveBar
    isInProgress={false}
    actions={actions}
  />
)

export const Approved = () => (
  <StepInCard
    installs={uatInstalls}
    groupName="uat"
    hasResponse
    responseType="approve"
    showApproveBar={false}
    isInProgress={false}
  />
)

export const Skipped = () => (
  <StepInCard
    installs={uatInstalls}
    groupName="uat"
    hasResponse
    responseType="skip"
    showApproveBar={false}
    isInProgress={false}
  />
)

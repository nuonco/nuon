export default {
  title: 'Workflows/VerifyHealthStepDetails',
}

import type { TWorkflowStep } from '@/types'
import { VerifyHealthStepDetails } from './VerifyHealthStepDetails'

const checks = [
  { kind: 'Deployment', name: 'whoami', health: 'healthy' },
  { kind: 'Service', name: 'whoami', health: 'healthy' },
  { kind: 'HTTPProbe', name: 'public-endpoint', health: 'healthy' },
  {
    kind: 'ExecProbe',
    name: 'gate-test-always-fails',
    health: 'unhealthy',
    message: 'exit code 1',
  },
  {
    kind: 'HTTPProbe',
    name: 'unresolvable-target',
    health: 'unknown',
    message: 'probe target could not be resolved from install state yet',
  },
  {
    kind: 'HTTPProbe',
    name: 'old-endpoint',
    health: 'unhealthy',
    message: 'this probe was deleted from the config',
    removed: true,
  },
]

const holdingStep = {
  id: 'step-1',
  owner_id: 'inl123',
  finished: false,
  status: {
    status: 'in-progress',
    status_human_description:
      'healthy for 25s of the 1m0s window — 5 of 6 resources healthy, 1 could not be checked',
    created_at_ts: 1785343200,
    metadata: { checks },
    history: [
      {
        status: 'in-progress',
        status_human_description:
          'waiting for a healthy verdict (currently progressing) — Deployment whoami/whoami: waiting for rollout',
        created_at_ts: 1785343140,
      },
    ],
  },
} as unknown as TWorkflowStep

const failedStep = {
  id: 'step-2',
  owner_id: 'inl123',
  finished: true,
  status: {
    status: 'error',
    status_human_description:
      'component is unhealthy: ExecProbe gate-test-always-fails: exit code 1',
    created_at_ts: 1785343260,
    history: [
      {
        status: 'in-progress',
        status_human_description:
          'component is unhealthy: ExecProbe gate-test-always-fails: exit code 1',
        created_at_ts: 1785343250,
        metadata: { checks },
      },
    ],
  },
} as unknown as TWorkflowStep

export const Holding = () => (
  <VerifyHealthStepDetails step={holdingStep} orgId="org1" componentId="cmp1" />
)

export const FailedLockedSnapshot = () => (
  <VerifyHealthStepDetails step={failedStep} orgId="org1" componentId="cmp1" />
)

export const NoChecksYet = () => (
  <VerifyHealthStepDetails
    step={{ id: 's', owner_id: 'inl123', status: { status: 'in-progress' } } as unknown as TWorkflowStep}
  />
)

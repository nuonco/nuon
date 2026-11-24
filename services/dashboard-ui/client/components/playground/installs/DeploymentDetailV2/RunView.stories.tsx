export default {
  title: 'Playground / Installs / Run view',
}

import type { TWorkflowStep } from '@/types'
import { SimulatedRunView } from './SimulatedRun'
import { StepTooltipContent } from './StepTooltip'

export const Running = () => <SimulatedRunView />

const plan = {
  id: 'step-plan',
  name: 'plan acme-api',
  execution_type: 'approval',
  finished: true,
  status: {
    status: 'success',
    metadata: {
      deny_violations: [{ policy_id: 'pol-1', message: 'Denied by policy' }],
      warn_violations: [
        { policy_id: 'pol-2', message: 'Warning one' },
        { policy_id: 'pol-3', message: 'Warning two' },
      ],
      skip_findings: [{ policy_id: 'pol-4', message: 'Skipped' }],
    },
  },
  metadata: { component_name: 'acme-api' },
  approval: {
    id: 'apr-1',
    type: 'helm_approval',
    changes_create: 1,
    changes_update: 3,
    changes_delete: 0,
    changes_replace: 1,
    response: { id: 'res-1', type: 'approve' },
  },
} as TWorkflowStep

const apply = {
  id: 'step-apply',
  name: 'apply acme-api',
  execution_type: 'system',
  finished: false,
  status: {
    status: 'in-progress',
    status_human_description: 'Waiting for rollout to finish',
  },
  metadata: { component_name: 'acme-api' },
} as TWorkflowStep

const retried = {
  id: 'step-retried',
  name: 'apply acme-worker',
  execution_type: 'system',
  retried: true,
  finished: true,
  status: {
    status: 'error',
    status_human_description: 'Helm upgrade failed',
  },
  metadata: { component_name: 'acme-worker' },
} as TWorkflowStep

const clean = {
  id: 'step-clean',
  name: 'Sync install configuration',
  execution_type: 'system',
  finished: true,
  status: { status: 'success' },
} as TWorkflowStep

const steps = [clean, plan, apply, retried]

export const StepTooltips = () => (
  <div className="flex flex-wrap items-start gap-6 p-6">
    {steps.map((step) => (
      <div key={step.id} className="rounded-md border bg-background px-2 py-1">
        <StepTooltipContent step={step} steps={steps} />
      </div>
    ))}
  </div>
)

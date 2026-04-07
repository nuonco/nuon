export default {
  title: 'Workflows/StepDetails/SandboxRunApply',
}

import { SandboxRunApply, SandboxRunApplySkeleton, SandboxRunLogsSkeleton } from './SandboxRunApply'
import type { TWorkflowStep, TSandboxRun } from '@/types'

const mockStep: TWorkflowStep = {
  id: 'step-1',
  status: 'active',
} as TWorkflowStep

const mockSandboxRun: TSandboxRun = {
  id: 'sanrun-abc123',
  status_v2: {
    status: 'active',
    status_human_description: 'Running sandbox',
  },
  log_stream: null,
} as TSandboxRun

export const Active = () => (
  <SandboxRunApply step={mockStep} sandboxRun={mockSandboxRun} />
)

export const Error = () => (
  <SandboxRunApply
    step={mockStep}
    sandboxRun={{
      ...mockSandboxRun,
      status_v2: { status: 'error', status_human_description: 'Sandbox run failed' },
    } as TSandboxRun}
  />
)

export const Null = () => (
  <SandboxRunApply step={mockStep} sandboxRun={null as any} />
)

export const ApplySkeleton = () => <SandboxRunApplySkeleton />

export const LogsSkeleton = () => <SandboxRunLogsSkeleton />

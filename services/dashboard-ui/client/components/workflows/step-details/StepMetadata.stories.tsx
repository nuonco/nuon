export default {
  title: 'Workflows/StepMetadata',
}

import { StepMetadata } from './StepMetadata'
import type { TWorkflowStep } from '@/types'

const mockStep = {
  id: 'step-1',
  name: 'deploy component',
  created_by: { email: 'user@example.com' },
  status: {
    status: 'success',
    history: [
      { status: 'pending', created_at_ts: 1704067200 },
      { status: 'in-progress', created_at_ts: 1704067260 },
    ],
  },
} as TWorkflowStep

export const Default = () => <StepMetadata step={mockStep} />

export default {
  title: 'Workflows/PolicyViolations',
}

import { PolicyViolations } from './PolicyViolations'
import type { TWorkflowStep } from '@/types'

const violationStep = {
  id: 'step-1',
  policy_results: {
    results: [
      { action: 'deny', message: 'Resource exceeds cost limit', policy_id: 'pol-1' },
      { action: 'deny', message: 'Missing required tags', policy_id: 'pol-2' },
      { action: 'warn', message: 'Consider using smaller instance', policy_id: 'pol-3' },
    ],
  },
} as TWorkflowStep

const passedStep = {
  id: 'step-2',
  policy_results: {
    results: [{ action: 'allow' }],
  },
} as TWorkflowStep

export const WithViolations = () => <PolicyViolations step={violationStep} />

export const AllPassed = () => <PolicyViolations step={passedStep} />

export default {
  title: 'Workflows/PolicyCountsBadge',
}

import { PolicyCountsBadge } from './PolicyCountsBadge'
import type { TWorkflowStep } from '@/types'

const passedStep = {
  id: 'step-1',
  policy_results: {
    results: [{ action: 'allow' }],
  },
} as TWorkflowStep

const violationStep = {
  id: 'step-2',
  policy_results: {
    results: [
      { action: 'deny', message: 'Denied' },
      { action: 'warn', message: 'Warning' },
    ],
  },
} as TWorkflowStep

const noPolicyStep = {
  id: 'step-3',
} as TWorkflowStep

export const Passed = () => <PolicyCountsBadge step={passedStep} />

export const WithViolations = () => <PolicyCountsBadge step={violationStep} />

export const NoPolicy = () => <PolicyCountsBadge step={noPolicyStep} />

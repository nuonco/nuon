import { expect, test } from 'bun:test'
import type { TWorkflow } from '@/types'
import { deploymentV2TabOrder, isWorkflowRunning } from './tabs'

const workflow = (status: string, type = 'app_branch_config_update') =>
  ({ type, status: { status } }) as TWorkflow

test.each([
  'in-progress',
  'approval-awaiting',
  'pending',
  'failed-pending-retry',
])('opens the run view first while the workflow is %s', (status) => {
  expect(isWorkflowRunning(workflow(status))).toBe(true)
  expect(deploymentV2TabOrder(workflow(status))).toEqual([
    'live',
    'template',
    'workflow',
    'changes',
  ])
})

test.each(['success', 'error', 'cancelled'])(
  'opens the change summary first once the workflow is %s',
  (status) => {
    expect(isWorkflowRunning(workflow(status))).toBe(false)
    expect(deploymentV2TabOrder(workflow(status))).toEqual([
      'changes',
      'template',
      'workflow',
      'live',
    ])
  }
)

test('workflows without a template tab keep the same ordering rule', () => {
  expect(deploymentV2TabOrder(workflow('in-progress', 'input_update'))).toEqual(
    ['live', 'workflow', 'changes']
  )
  expect(deploymentV2TabOrder(workflow('success', 'input_update'))).toEqual([
    'changes',
    'workflow',
    'live',
  ])
})

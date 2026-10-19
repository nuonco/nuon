import { describe, expect, test } from 'bun:test'
import { isWorkflowPath, workflowPageSuffix } from './workflow-href'

describe('workflowPageSuffix', () => {
  test('links the workflow name to the workflow page', () => {
    expect(workflowPageSuffix('wf-1', false)).toBe('/workflows/wf-1')
    expect(workflowPageSuffix('wf-1', true)).toBe('/workflows/wf-1/workflow')
  })

  test('links view details to that page with the step panel', () => {
    expect(workflowPageSuffix('wf-1', true, 'step 1')).toBe(
      '/workflows/wf-1/workflow?panel=step%201'
    )
  })
})

describe('isWorkflowPath', () => {
  test('matches the workflow page and not a branch panel', () => {
    expect(isWorkflowPath('/org/installs/ins/workflows/wf-1', 'wf-1')).toBe(true)
    expect(
      isWorkflowPath('/org/apps/app/installs/ins/deployments/wf-1/workflow', 'wf-1')
    ).toBe(true)
    expect(
      isWorkflowPath('/org/apps/app/branches/br?panel=install-group', 'wf-1')
    ).toBe(false)
  })
})

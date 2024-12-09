import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { postWorkflowRun } from './post-workflow-run'

const installId = 'install-id'
const orgId = 'org-id'
const workflowConfigId = 'workflow-config-id'

test('postWorkflowRun should return an new install action workflow run object', async () => {
  const spec = await postWorkflowRun({ installId, orgId, workflowConfigId })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('status')
  expect(spec).toHaveProperty("trigger_type")
})

test('postWorkflowRun should throw an error when it can not kick off an install action workflow run object', async () => {
  try {
    await postWorkflowRun({ installId, orgId, workflowConfigId })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to kick off an action workflow]`)
  }
})

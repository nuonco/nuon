import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getInstallWorkflowRun } from './get-install-workflow-run'

const orgId = 'org-id'
const installId = 'install-id'
const workflowRunId = 'workflow-id'

test('getInstallWorkflowRun should return a install action workflow run object', async () => {
  const spec = await getInstallWorkflowRun({
    installId,
    orgId,
    workflowRunId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('status')

})

test('getInstallWorkflowRun should throw an error when it can not find an install action workflow run', async () => {
  try {
    await getInstallWorkflowRun({
      installId,
      orgId,
      workflowRunId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Failed to fetch install action workflow run]`
    )
  }
})

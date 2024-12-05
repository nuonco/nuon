import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getWorkflow } from './get-workflow'

const orgId = 'org-id'
const workflowId = 'workflow-id'

test('getWorkflow should return a action workflow object', async () => {
  const spec = await getWorkflow({
    orgId,
    workflowId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('name')
  expect(spec).toHaveProperty('configs')
})

test('getWorkflow should throw an error when it can not find an action workflow', async () => {
  try {
    await getWorkflow({
      orgId,
      workflowId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Failed to fetch action workflow]`
    )
  }
})

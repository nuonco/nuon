import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppActionWorkflow } from './get-app-action-workflow'

describe('getAppActionWorkflow should handle response status codes from GET apps/:id/action-workflows endpoint', () => {
  const orgId = 'test-id'
  const actionWorkflowId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppActionWorkflow({ actionWorkflowId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('configs')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppActionWorkflow({ actionWorkflowId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
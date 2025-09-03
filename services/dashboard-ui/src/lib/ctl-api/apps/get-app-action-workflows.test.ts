import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppActionWorkflows } from './get-app-action-workflows'

describe('getAppActionWorkflows should handle response status codes from GET apps/:id/action-workflows endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppActionWorkflows({ appId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('configs')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppActionWorkflows({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
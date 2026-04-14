import { describe, expect, test } from 'vitest'
import { getComponentBuild } from './get-component-build'

describe('getComponentBuild should handle response status codes from GET components/:componentId/builds/:buildId endpoint', () => {
  const componentId = 'test-component-id'
  const buildId = 'test-build-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getComponentBuild({ componentId, buildId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
  }, 60000)
})

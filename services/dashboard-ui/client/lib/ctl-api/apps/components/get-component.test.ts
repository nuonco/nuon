import { describe, expect, test } from 'vitest'
import { getComponent } from './get-component'

describe('getComponent should handle response status codes from GET components/:componentId endpoint', () => {
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getComponent({ componentId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
  }, 60000)
})

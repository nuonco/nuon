import { describe, expect, test } from 'vitest'
import { getComponentBuilds } from './get-component-builds'

describe('getComponentBuilds should handle response status codes from GET /builds?:params endpoint', () => {
  const orgId = 'test-id'
  const componentId = 'test-id'

  test('200 status', async () => {
    const result = await getComponentBuilds({ componentId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result.data)).toBe(true)
  })
})

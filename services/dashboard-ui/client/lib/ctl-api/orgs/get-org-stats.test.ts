import { describe, expect, test } from 'vitest'
import { getOrgStats } from './get-org-stats'

describe('getOrgStats should handle response status codes from GET orgs/current/stats endpoint', () => {
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getOrgStats({ orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
  })
})

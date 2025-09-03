import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponentBuild } from './get-component-build'

describe('getComponentBuild should handle response status codes from GET builds/:id endpoint', () => {
  const orgId = 'test-id'
  const buildId = 'test-id'
  test('200 status', async () => {
    const spec = await getComponentBuild({ buildId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('status')
  }, 60000)

  test.each(badResponseCodes)('%s status', async () => {
    await getComponentBuild({ buildId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

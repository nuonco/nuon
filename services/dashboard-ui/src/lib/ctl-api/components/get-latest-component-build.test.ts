import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getLatestComponentBuild } from './get-latest-component-build'

describe('getLatestComponentBuild should handle response status codes from GET components/:id/builds/latest endpoint', () => {
  const orgId = 'test-id'
  const componentId = 'test-id'
  test('200 status', async () => {
    const spec = await getLatestComponentBuild({ componentId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getLatestComponentBuild({ componentId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
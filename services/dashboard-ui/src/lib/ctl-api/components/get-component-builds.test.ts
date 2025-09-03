import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponentBuilds } from './get-component-builds'

describe.skip('getComponentBuilds should handle response status codes from GET components/:id/builds endpoint', () => {
  const orgId = 'test-id'
  const componentId = 'test-id'
  test('200 status', async () => {
    const spec = await getComponentBuilds({ componentId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('status')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getComponentBuilds({ componentId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

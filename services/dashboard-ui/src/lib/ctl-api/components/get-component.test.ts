import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponent } from './get-component'

describe('getComponent should handle response status codes from GET components/:id endpoint', () => {
  const orgId = 'test-id'
  const componentId = 'test-id'
  test('200 status', async () => {
    const spec = await getComponent({ componentId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getComponent({ componentId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
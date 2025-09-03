import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstall } from './get-install'

describe('getInstall should handle response status codes from GET installs/:id endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  test('200 status', async () => {
    const spec = await getInstall({ installId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
    expect(spec).toHaveProperty('composite_component_status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getInstall({ installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
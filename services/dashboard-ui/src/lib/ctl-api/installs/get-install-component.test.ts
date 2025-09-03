import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallComponent } from './get-install-component'

describe('getInstallComponent should handle response status codes from GET installs/:id/components/:id endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  const componentId = 'test-id'
  test('200 status', async () => {
    const spec = await getInstallComponent({ componentId, installId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getInstallComponent({ componentId, installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallComponents } from './get-install-components'

describe.skip('getInstallComponents should handle response status codes from GET installs/:id/components endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  test('200 status', async () => {
    const spec = await getInstallComponents({ installId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('status')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getInstallComponents({ installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })

})

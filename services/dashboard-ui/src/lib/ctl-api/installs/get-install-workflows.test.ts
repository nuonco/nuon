import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallWorkflows } from './get-install-workflows'

describe('getInstallWorkflows should handle response status codes from GET installs/:id/workflows endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  test('200 status', async () => {
    const spec = await getInstallWorkflows({ installId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('status')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getInstallWorkflows({ installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
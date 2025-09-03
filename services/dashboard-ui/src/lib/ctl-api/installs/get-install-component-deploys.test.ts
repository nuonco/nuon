import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallComponentDeploys } from './get-install-component-deploys'

describe('getInstallComponentDeploys should handle response status codes from GET installs/:id/components/:id/deploys endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  const componentId = 'test-id'
  test('200 status', async () => {
    const spec = await getInstallComponentDeploys({ componentId, installId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('status')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getInstallComponentDeploys({ componentId, installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
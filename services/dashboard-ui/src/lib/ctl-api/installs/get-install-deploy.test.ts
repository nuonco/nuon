import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallDeploy } from './get-install-deploy'

describe('getInstallDeploy should handle response status codes from GET installs/:id/deploys/:id endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  const installDeployId = 'test-id'
  test('200 status', async () => {
    const spec = await getInstallDeploy({ installDeployId, installId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getInstallDeploy({ installDeployId, installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
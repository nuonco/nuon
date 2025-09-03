import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getVCSConnections } from './get-vcs-connections'

describe('getVCSConnections should handle response status codes from GET vcs/connections endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const spec = await getVCSConnections({ orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('github_install_id')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getVCSConnections({ orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
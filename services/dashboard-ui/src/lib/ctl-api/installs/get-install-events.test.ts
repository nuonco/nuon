import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallEvents } from './get-install-events'

// NOTE(nnnnat): this has been replaced by workflows
describe.skip('getInstallEvents should handle response status codes from GET installs/:id/events endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  test('200 status', async () => {
    const spec = await getInstallEvents({ installId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('event_type')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getInstallEvents({ installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { deployComponents } from './deploy-components'

describe('deployComponents should handle response status codes from POST installs/:id/components/deploy-all endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  test('200 status', async () => {
    const spec = await deployComponents({ installId, orgId })
    expect(spec).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await deployComponents({ installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
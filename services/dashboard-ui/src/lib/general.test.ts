import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { joinWaitlist } from './general'

describe('joinWaitlist should handle response status codes from POST general/waitlist endpoint', () => {
  test('200 status', async () => {
    const spec = await joinWaitlist({ org_name: 'test-name' })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('org_name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await joinWaitlist({ org_name: 'test-name' }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

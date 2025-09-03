import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrgRunnerGroup } from './get-org-runner-group'

describe('getOrgRunnerGroup should handle response status codes from GET orgs/current/runner-group endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const spec = await getOrgRunnerGroup({ orgId })
    expect(spec).toHaveProperty("runners")
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getOrgRunnerGroup({ orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

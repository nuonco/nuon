import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { createComponentBuild } from './create-component-build'

describe('createComponentBuild should handle response status codes from POST components/:id/builds endpoint', () => {
  const orgId = 'test-id'
  const componentId = 'test-id'
  test('200 status', async () => {
    const spec = await createComponentBuild({ componentId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await createComponentBuild({ componentId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
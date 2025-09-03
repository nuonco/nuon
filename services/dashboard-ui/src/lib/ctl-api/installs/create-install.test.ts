import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { createInstall } from './create-install'

describe('createInstall should handle response status codes from POST apps/:id/installs endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  const data = { name: 'test-install' }
  test('200 status', async () => {
    const spec = await createInstall({ appId, orgId, data })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
    expect(spec).toHaveProperty('composite_component_status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await createInstall({ appId, orgId, data }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
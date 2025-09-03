import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { updateInstall } from './update-install'

describe('updateInstall should handle response status codes from PATCH installs/:id endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-id'
  const data = { name: 'updated-name' }
  test('200 status', async () => {
    const spec = await updateInstall({ data, installId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
    expect(spec).toHaveProperty('composite_component_status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await updateInstall({ data, installId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
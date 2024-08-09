import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getAppInstalls } from './get-app-installs'

const appId = 'app-id'
const orgId = 'org-id'

test('getAppInstalls should return an array of install object', async () => {
  const spec = await getAppInstalls({
    appId,
    orgId,
  })

  expect(spec).toHaveLength(3)
  spec.map((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('name')
  })
})

test('getAppInstalls should throw an error when it can not find any installs for the app', async () => {
  try {
    await getAppInstalls({
      appId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch app installs]`)
  }
})

import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getAppComponents } from './get-app-components'

const appId = 'app-id'
const orgId = 'org-id'

test('getAppComponents should return an array of component object', async () => {
  const spec = await getAppComponents({
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

test('getAppComponents should throw an error when it can not find any components for the app', async () => {
  try {
    await getAppComponents({
      appId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch app components]`)
  }
})

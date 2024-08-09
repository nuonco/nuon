import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getApps } from './get-apps'

const orgId = 'org-id'

test('getApps should return an array of app object', async () => {
  const spec = await getApps({
    orgId,
  })

  expect(spec).toHaveLength(3)
  spec.map((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('name')
    expect(s).toHaveProperty('cloud_platform')
  })
})

test('getApps should throw an error when it can not find any apps', async () => {
  try {
    await getApps({
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch apps]`)
  }
})

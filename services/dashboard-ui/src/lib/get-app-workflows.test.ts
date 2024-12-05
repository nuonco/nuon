import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getAppWorkflows } from './get-app-workflows'

const orgId = 'org-id'
const appId = 'app-id'

test('getAppWorkflows should return an array of app workflow object', async () => {
  const spec = await getAppWorkflows({
    appId,
    orgId,
  })

  expect(spec).toHaveLength(3)
  spec.map((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('name')
    expect(s).toHaveProperty('configs')
  })
})

test('getAppWorkflows should throw an error when it can not find any app workflows', async () => {
  try {
    await getAppWorkflows({
      appId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch app action workflows]`)
  }
})

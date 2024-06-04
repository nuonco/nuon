import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getOrg } from './get-org'

const orgId = 'org-id'

test('getOrg should return a org object', async () => {
  const spec = await getOrg({
    orgId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('name')
})

test('getOrg should throw an error when it can not find an org', async () => {
  try {
    await getOrg({
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch current org]`)
  }
})

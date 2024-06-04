import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getOrgs } from './get-orgs'

test('getOrgs should return an array of install object', async () => {
  const spec = await getOrgs()

  expect(spec).toHaveLength(9)
  spec.forEach((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('name')
  })
})

test('getOrgs should throw an error when it can not find orgs', async () => {
  try {
    await getOrgs()
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch orgs]`)
  }
})

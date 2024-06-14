import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { postOrg } from './post-org'

test('postOrg should return an new org object', async () => {
  const spec = await postOrg({ name: 'test' })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('name')
})

test('postOrg should throw an error when it can not find orgs', async () => {
  try {
    await postOrg({ name: 'test' })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to create org]`)
  }
})

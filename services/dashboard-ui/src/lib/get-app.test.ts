import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getApp } from './get-app'

const appId = 'app-id'
const orgId = 'org-id'

test('getApp should return a app object', async () => {
  const spec = await getApp({
    appId,
    orgId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('name')
  expect(spec).toHaveProperty('cloud_platform')
})

test('getApp should throw an error when it can not find a app', async () => {
  try {
    await getApp({
      appId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch app]`)
  }
})

import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getBuild } from './get-build'

const buildId = 'build-id'
const orgId = 'org-id'

test('getBuild should return a build object', async () => {
  const spec = await getBuild({
    buildId,
    orgId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('component_id')
  expect(spec).toHaveProperty('component_config_connection_id')
})

test('getBuild should throw an error when it can not find a build', async () => {
  try {
    await getBuild({
      buildId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch build]`)
  }
})

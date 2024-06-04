import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getInstalls } from './get-installs'

const orgId = 'org-id'

test('getInstalls should return an array of install object', async () => {
  const spec = await getInstalls({
    orgId,
  })

  expect(spec).toHaveLength(9)
  spec.forEach((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('name')
  })
})

test('getInstalls should throw an error when it can not find installs', async () => {
  try {
    await getInstalls({
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch installs]`)
  }
})

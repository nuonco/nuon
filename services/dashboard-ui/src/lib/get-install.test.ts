import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getInstall } from './get-install'

const installId = 'install-id'
const orgId = 'org-id'

test('getInstall should return a install object', async () => {
  const spec = await getInstall({
    installId,
    orgId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('name')
})

test('getInstall should throw an error when it can not find a install', async () => {
  try {
    await getInstall({
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch install]`)
  }
})

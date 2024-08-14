import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getVCSConnections } from './get-vcs-connections'

const orgId = 'org-id'

test('getVCSConnections should return an array of install object', async () => {
  const spec = await getVCSConnections({
    orgId,
  })

  expect(spec).toHaveLength(3)
  spec.forEach((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('github_install_id')
  })
})

test('getVCSConnections should throw an error when it can not find VCS connections', async () => {
  try {
    await getVCSConnections({
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch VCS connections]`)
  }
})

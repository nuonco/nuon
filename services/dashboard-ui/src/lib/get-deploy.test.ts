import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getDeploy } from './get-deploy'

const deployId = 'deploy-id'
const installId = 'install-id'
const orgId = 'org-id'

test('getDeploy should return a deploy object', async () => {
  const spec = await getDeploy({
    deployId,
    installId,
    orgId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('component_id')
  expect(spec).toHaveProperty('install_id')
})

test('getDeploy should throw an error when it can not find deploy', async () => {
  try {
    await getDeploy({
      deployId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch deploy]`)
  }
})

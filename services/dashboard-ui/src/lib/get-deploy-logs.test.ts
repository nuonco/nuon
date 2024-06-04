import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getDeployLogs } from './get-deploy-logs'

const deployId = 'deploy-id'
const installId = 'install-id'
const orgId = 'org-id'

test('getDeployLogs should return an array of deploy log objects', async () => {
  const spec = await getDeployLogs({
    deployId,
    installId,
    orgId,
  })

  expect(spec).toMatchInlineSnapshot(`
    [
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
    ]
  `)
})

test('getDeployLogs should throw an error when it can not find deploy logs', async () => {
  try {
    await getDeployLogs({
      deployId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch deploy logs]`)
  }
})

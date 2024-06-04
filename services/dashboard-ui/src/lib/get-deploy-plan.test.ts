import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getDeployPlan } from './get-deploy-plan'

const deployId = 'deploy-id'
const installId = 'install-id'
const orgId = 'org-id'

test('getDeployPlan should return a deploy plan object', async () => {
  const spec = await getDeployPlan({
    deployId,
    installId,
    orgId,
  })

  expect(spec?.actual).toBeNull()
})

test('getDeployPlan should throw an error when it can not find deploy plan', async () => {
  try {
    await getDeployPlan({
      deployId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch deploy plan]`)
  }
})

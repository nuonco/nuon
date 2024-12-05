import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getInstallWorkflowRuns } from './get-install-workflow-runs'

const orgId = 'org-id'
const installId = 'install-id'

test('getInstallWorkflowRuns should return an array of app object', async () => {
  const spec = await getInstallWorkflowRuns({
    installId,
    orgId,
  })

  expect(spec).toHaveLength(3)
  spec.map((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('status')
    expect(s).toHaveProperty('status_description')
  })
})

test('getInstallWorkflowRuns should throw an error when it can not find any app workflows', async () => {
  try {
    await getInstallWorkflowRuns({
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch install action workflow runs]`)
  }
})

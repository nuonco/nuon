import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getInstallRunnerGroup } from './get-install-runner-group'

const orgId = 'org-id'
const installId = 'install-id'

test('getInstallRunnerGroup should return a install runner group', async () => {
  const spec = await getInstallRunnerGroup({
    installId,
    orgId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
  expect(spec).toHaveProperty('runners')

})

test('getInstallRunnerGroup should throw an error when it can not find an install runner group', async () => {
  try {
    await getInstallRunnerGroup({
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Failed to fetch install runner group]`
    )
  }
})

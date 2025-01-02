import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getInstallReadme } from './get-install-readme'

const orgId = 'org-id'
const installId = 'install-id'

test('getInstallReadme should return a runner object', async () => {
  const spec = await getInstallReadme({
    orgId,
    installId,
  })

  expect(spec).toHaveProperty('readme')
})

test('getInstallReadme should throw an error when it can not find a runner', async () => {
  try {
    await getInstallReadme({
      orgId,
      installId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch install readme]`)
  }
})

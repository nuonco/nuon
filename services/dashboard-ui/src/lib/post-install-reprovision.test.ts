import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { postInstallReprovision } from './post-install-reprovision'

const installId = "install-id"
const orgId = "org-id"

test('postInstallReprovision should return an string when install reprovision is kicked off', async () => {
  const spec = await postInstallReprovision({ orgId, installId })
  expect(spec).not.toBeNull()
})

test('postInstallReprovision should throw an error when install reprovision fails to kick off', async () => {
  try {
    await postInstallReprovision({ orgId, installId })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to kick off reprovision]`)
  }
})

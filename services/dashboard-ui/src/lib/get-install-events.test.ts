import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getInstallEvents } from './get-install-events'

const installId = 'install-id'
const orgId = 'org-id'

test('getInstallEvents should return an array of install event objects', async () => {
  const spec = await getInstallEvents({
    installId,
    orgId,
  })

  expect(spec).toHaveLength(9)
  spec.forEach((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('payload')
    expect(s).toHaveProperty('operation')
  })
})

test('getInstallEvents should throw an error when it can not find install events', async () => {
  try {
    await getInstallEvents({
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Failed to fetch install events]`
    )
  }
})

import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getSandboxRunLogs } from './get-sandbox-run-logs'

const runId = 'run-id'
const installId = 'install-id'
const orgId = 'org-id'

test('getSandboxRunLogs should return an array of run log objects', async () => {
  const spec = await getSandboxRunLogs({
    runId,
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

test('getSandboxRunLogs should throw an error when it can not find run logs', async () => {
  try {
    await getSandboxRunLogs({
      runId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Failed to fetch sandbox run logs]`
    )
  }
})

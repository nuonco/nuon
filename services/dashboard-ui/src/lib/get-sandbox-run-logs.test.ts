import { afterAll, expect, test, vi } from 'vitest'
import { getSandboxRunLogs } from './get-sandbox-run-logs'

const runId = 'run-id'
const installId = 'install-id'
const orgId = 'org-id'
const log = { State: { current: 'test' } }

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve([log])),
  })
  .mockResolvedValueOnce({
    ok: false,
    json: () => new Promise((resolve) => resolve('error')),
  })

vi.mock('../utils', async (og) => {
  const mod = await og<typeof import('../utils')>()
  return {
    ...mod,
    getFetchOpts: vi.fn().mockResolvedValue({
      cache: 'no-store',
      headers: {
        Authorization: 'Bearer test-token',
        'Content-Type': 'application/json',
        'X-Nuon-Org-ID': 'org-id',
      },
    }),
  }
})

afterAll(() => {
  vi.restoreAllMocks()
})

test('getSandboxRunLogs should return an array of run log objects', async () => {
  const spec = await getSandboxRunLogs({
    runId,
    installId,
    orgId,
  })

  expect(spec).toContain(log)
  expect(fetch).toBeCalledWith(
    'https://ctl.prod.nuon.co/v1/installs/install-id/sandbox-run/run-id/logs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getSandboxRunLogs should throw an error when it can not find run logs', async () => {
  try {
    await getSandboxRunLogs({
      runId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://ctl.prod.nuon.co/v1/installs/install-id/sandbox-run/run-id/logs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

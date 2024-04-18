import { afterAll, expect, test, vi } from 'vitest'
import { getSandboxRun } from './get-sandbox-run'

const runId = 'run-id'
const installId = 'install-id'
const orgId = 'org-id'
const run = {
  id: runId,
  install_id: installId,
}

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve([run])),
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

test('getSandboxRun should return a sandbox run object', async () => {
  const spec = await getSandboxRun({
    runId,
    installId,
    orgId,
  })

  expect(spec).toEqual(run)
  expect(fetch).toBeCalledWith(
    'https://ctl.prod.nuon.co/v1/installs/install-id/sandbox-runs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getSandboxRun should throw an error when it can not find a sandbox run', async () => {
  try {
    await getSandboxRun({
      runId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://ctl.prod.nuon.co/v1/installs/install-id/sandbox-runs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

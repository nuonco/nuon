import { afterAll, expect, test, vi } from 'vitest'
import { getBuildLogs } from './get-build-logs'

const buildId = 'build-id'
const componentId = 'component-id'
const orgId = 'org-id'

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve([])),
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

test('getBuildLogs should return a build object', async () => {
  const spec = await getBuildLogs({
    componentId,
    buildId,
    orgId,
  })

  expect(spec).toHaveLength(0)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/components/component-id/builds/build-id/logs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': 'org-id',
      }),
    })
  )
})

test('getBuildLogs should throw an error when it can not find a build', async () => {
  try {
    await getBuildLogs({
      componentId,
      buildId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/components/component-id/builds/build-id/logs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': 'org-id',
      }),
    })
  )
})

import { afterAll, expect, test, vi } from 'vitest'
import { getBuild } from './get-build'

const buildId = 'build-id'
const componentId = 'component-id'
const orgId = 'org-id'
const build = { id: buildId, component_id: componentId }

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve(build)),
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

test('getBuild should return a build object', async () => {
  const spec = await getBuild({
    buildId,
    orgId,
  })

  expect(spec).toEqual(build)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/components/builds/build-id',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': 'org-id',
      }),
    })
  )
})

test('getBuild should throw an error when it can not find a build', async () => {
  try {
    await getBuild({
      buildId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/components/builds/build-id',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': 'org-id',
      }),
    })
  )
})

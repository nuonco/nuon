import { afterAll, expect, test, vi } from 'vitest'
import { getBuildPlan } from './get-build-plan'

const buildId = 'build-id'
const componentId = 'component-id'
const orgId = 'org-id'

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve({})),
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

test('getBuildPlan should return a build object', async () => {
  const spec = await getBuildPlan({
    componentId,
    buildId,
    orgId,
  })

  expect(spec).toEqual({})
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/components/component-id/builds/build-id/plan',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': 'org-id',
      }),
    })
  )
})

test('getBuildPlan should throw an error when it can not find a build', async () => {
  try {
    await getBuildPlan({
      componentId,
      buildId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch build plan]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/components/component-id/builds/build-id/plan',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': 'org-id',
      }),
    })
  )
})

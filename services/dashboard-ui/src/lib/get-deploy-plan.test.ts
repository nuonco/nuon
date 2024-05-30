import { afterAll, expect, test, vi } from 'vitest'
import { getDeployPlan } from './get-deploy-plan'

const deployId = 'deploy-id'
const installId = 'install-id'
const orgId = 'org-id'
const plan = { actual: "test" }

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve(plan)),
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

test('getDeployPlan should return a deploy plan object', async () => {
  const spec = await getDeployPlan({
    deployId,
    installId,
    orgId,
  })

  expect(spec).toEqual(plan)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs/install-id/deploys/deploy-id/plan',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getDeployPlan should throw an error when it can not find deploy plan', async () => {
  try {
    await getDeployPlan({
      deployId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch deploy plan]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs/install-id/deploys/deploy-id/plan',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

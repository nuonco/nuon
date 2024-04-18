import { afterAll, expect, test, vi } from 'vitest'
import { getDeploy } from './get-deploy'

const deployId = 'deploy-id'
const installId = 'install-id'
const orgId = 'org-id'
const deploy = { id: deployId }

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve(deploy)),
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

test('getDeploy should return a deploy object', async () => {
  const spec = await getDeploy({
    deployId,
    installId,
    orgId,
  })

  expect(spec).toEqual(deploy)
  expect(fetch).toBeCalledWith(
    'https://ctl.prod.nuon.co/v1/installs/install-id/deploys/deploy-id',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getDeploy should throw an error when it can not find deploy', async () => {
  try {
    await getDeploy({
      deployId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://ctl.prod.nuon.co/v1/installs/install-id/deploys/deploy-id',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

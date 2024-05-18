import { afterAll, expect, test, vi } from 'vitest'
import { getInstall } from './get-install'

const installId = 'install-id'
const orgId = 'org-id'
const install = {
  id: installId,
  name: 'test',
}

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve(install)),
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

test('getInstall should return a install object', async () => {
  const spec = await getInstall({
    installId,
    orgId,
  })

  expect(spec).toEqual(install)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs/install-id',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getInstall should throw an error when it can not find a install', async () => {
  try {
    await getInstall({
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs/install-id',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

import { afterAll, expect, test, vi } from 'vitest'
import { getInstalls } from './get-installs'

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
    json: () => new Promise((resolve) => resolve([install])),
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

test('getInstalls should return an array of install object', async () => {
  const spec = await getInstalls({
    orgId,
  })

  expect(spec).toContain(install)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getInstalls should throw an error when it can not find installs', async () => {
  try {
    await getInstalls({
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch installs]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

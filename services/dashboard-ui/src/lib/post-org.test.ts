import { afterAll, expect, test, vi } from 'vitest'
import { postOrg } from './post-org'

const orgName = 'test org'
const orgId = 'org-id'
const org = {
  id: orgId,
  name: orgName,
}

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve(org)),
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
      },
    }),
  }
})

afterAll(() => {
  vi.restoreAllMocks()
})

test('postOrg should return an array of install object', async () => {
  const spec = await postOrg({ name: orgName })

  expect(spec).toEqual(org)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/orgs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
      }),
      method: 'POST',
      body: JSON.stringify({ name: orgName }),
    })
  )
})

test('postOrg should throw an error when it can not find orgs', async () => {
  try {
    await postOrg({ name: orgName })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/orgs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
      }),
    })
  )
})

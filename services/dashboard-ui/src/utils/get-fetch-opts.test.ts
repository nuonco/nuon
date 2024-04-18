import { afterAll, expect, test, vi } from 'vitest'
import { getFetchOpts } from './get-fetch-opts'

vi.mock('@auth0/nextjs-auth0', async (og) => {
  const mod = await og<typeof import('@auth0/nextjs-auth0')>()
  return {
    ...mod,
    getSession: vi
      .fn()
      .mockResolvedValueOnce({ accessToken: 'test-token' })
      .mockResolvedValueOnce(null),
  }
})

afterAll(() => {
  vi.restoreAllMocks()
})

test('getFetchOptions should return valid headers for the ctl api if the user is authenticated', async () => {
  const spec = await getFetchOpts('org-id')

  expect(spec).toHaveProperty('cache')
  expect(spec).toHaveProperty(
    'headers',
    expect.objectContaining({
      Authorization: 'Bearer test-token',
      'X-Nuon-Org-ID': 'org-id',
    })
  )
})

test('getFetchOptions should return headers with undefined token if the user is not authenticated', async () => {
  const spec = await getFetchOpts('org-id')

  expect(spec).toHaveProperty('cache')
  expect(spec).toHaveProperty(
    'headers',
    expect.objectContaining({
      Authorization: 'Bearer undefined',
      'X-Nuon-Org-ID': 'org-id',
    })
  )
})

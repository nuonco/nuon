import { afterAll, vi } from 'vitest'

vi.mock('../src/utils/get-fetch-opts', async (og) => {
  const mod = await og<typeof import('../src/utils/get-fetch-opts')>()
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

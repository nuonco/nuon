import { afterAll, expect, test, vi } from 'vitest'
import { getInstallEvents } from './get-install-events'

const eventId = 'event-id'
const installId = 'install-id'
const orgId = 'org-id'
const event = { id: eventId, operation: "test" }

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve([event])),
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

test('getInstallEvents should return an array of install event objects', async () => {
  const spec = await getInstallEvents({
    installId,
    orgId,
  })

  expect(spec).toContain(event)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs/install-id/events',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getInstallEvents should throw an error when it can not find install events', async () => {
  try {
    await getInstallEvents({
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs/install-id/events',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

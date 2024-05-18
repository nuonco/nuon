import { afterAll, expect, test, vi } from 'vitest'
import { getComponentConfig } from './get-component-config'

const componentId = 'component-id'
const orgId = 'org-id'
const config = { id: "test-id", component_id: componentId }

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve([config])),
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
        'X-Nuon-Org-ID': 'org-id'
      },
    }),
  }
})

afterAll(() => {
  vi.restoreAllMocks()
})

test('getComponentConfig should return a config object', async () => {
  const spec = await getComponentConfig({
    componentId,
    orgId,
  })

  expect(spec).toEqual(config)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/components/component-id/configs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getComponentConfig should throw an error when it can not find a config', async () => {
  try {
    await getComponentConfig({
      componentId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/components/component-id/configs',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

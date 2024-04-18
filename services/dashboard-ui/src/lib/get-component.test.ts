import { afterAll, expect, test, vi } from 'vitest'
import { getComponent } from './get-component'

const componentId = 'component-id'
const orgId = 'org-id'
const component = { id: componentId, name: 'test' }

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve(component)),
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

test('getComponent should return a component object', async () => {
  const spec = await getComponent({
    componentId,
    orgId,
  })

  expect(spec).toEqual(component)
  expect(fetch).toBeCalledWith(
    'https://ctl.prod.nuon.co/v1/components/component-id',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getComponent should throw an error when it can not find a component', async () => {
  try {
    await getComponent({
      componentId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://ctl.prod.nuon.co/v1/components/component-id',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

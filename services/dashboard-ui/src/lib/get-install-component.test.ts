import { afterAll, expect, test, vi } from 'vitest'
import { getInstallComponent } from './get-install-component'

const installComponentId = 'install-component-id'
const componentId = 'component-id'
const installId = 'install-id'
const orgId = 'org-id'
const installComponent = {
  id: installComponentId,
  component_id: componentId,
  component_name: 'test',
  install_id: installId,
}

global.fetch = vi
  .fn()
  .mockResolvedValueOnce({
    ok: true,
    json: () => new Promise((resolve) => resolve([installComponent])),
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

test('getInstallComponent should return a install component object', async () => {
  const spec = await getInstallComponent({
    installComponentId,
    installId,
    orgId,
  })

  expect(spec).toEqual(installComponent)
  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs/install-id/components',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

test('getInstallComponent should throw an error when it can not find a install component', async () => {
  try {
    await getInstallComponent({
      installComponentId,
      installId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch data]`)
  }

  expect(fetch).toBeCalledWith(
    'https://api.nuon.co/v1/installs/install-id/components',
    expect.objectContaining({
      headers: expect.objectContaining({
        Authorization: 'Bearer test-token',
        'X-Nuon-Org-ID': orgId,
      }),
    })
  )
})

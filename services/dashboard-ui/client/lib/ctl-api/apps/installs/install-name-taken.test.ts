import { afterEach, beforeEach, expect, test } from 'bun:test'
import { installNameTaken } from './install-name-taken'

type Page = { names: string[]; hasNext?: boolean }

// Only fetch is stubbed, so getAppInstalls, buildQueryParams and the paginated
// header handling in api() all run for real.
const pages: Page[] = []
const requests: URL[] = []

const pageResponse = ({ names, hasNext = false }: Page) =>
  new Response(JSON.stringify(names.map((name) => ({ name }))), {
    status: 200,
    headers: {
      'content-type': 'application/json',
      'X-Nuon-Page-Next': String(hasNext),
    },
  })

// fetch is swapped by hand rather than with spyOn, because mock.restore() in
// other test files also tears down global spies installed here.
const realFetch = globalThis.fetch
const stubFetch = (handler: () => Response) => {
  globalThis.fetch = Object.assign(
    async (input: Parameters<typeof fetch>[0]) => {
      requests.push(new URL(String(input), 'https://api.example.com'))
      return handler()
    },
    { preconnect: () => {} }
  ) as unknown as typeof fetch
}

beforeEach(() => {
  pages.length = 0
  requests.length = 0
  stubFetch(() => pageResponse(pages.shift() ?? { names: [] }))
})

afterEach(() => {
  globalThis.fetch = realFetch
})

test('searches the app installs endpoint by name and reports an exact match', async () => {
  pages.push({ names: ['staging-copy', 'staging'] })

  expect(
    await installNameTaken({ appId: 'app-1', orgId: 'org-1', name: 'staging' })
  ).toBe(true)

  expect(requests).toHaveLength(1)
  expect(requests[0].pathname).toBe('/v1/apps/app-1/installs')
  expect(requests[0].searchParams.get('q')).toBe('staging')
  expect(requests[0].searchParams.get('limit')).toBe('100')
})

test('substring matches alone do not make a name taken', async () => {
  pages.push({ names: ['staging-copy', 'pre-staging'] })

  expect(
    await installNameTaken({ appId: 'app-1', orgId: 'org-1', name: 'staging' })
  ).toBe(false)
})

test('keeps paging until the exact match is found', async () => {
  pages.push({ names: ['staging-a', 'staging-b'], hasNext: true })
  pages.push({ names: ['staging'] })

  expect(
    await installNameTaken({ appId: 'app-1', orgId: 'org-1', name: 'staging' })
  ).toBe(true)
  expect(requests).toHaveLength(2)
  expect(requests[1].searchParams.get('offset')).toBe('2')
})

test('an empty name is never taken and skips the lookup', async () => {
  expect(
    await installNameTaken({ appId: 'app-1', orgId: 'org-1', name: '   ' })
  ).toBe(false)
  expect(requests).toHaveLength(0)
})

test('a failed lookup surfaces as an error rather than an available name', async () => {
  stubFetch(
    () =>
      new Response(JSON.stringify({ error: 'boom' }), {
        status: 500,
        headers: { 'content-type': 'application/json' },
      })
  )

  await expect(
    installNameTaken({ appId: 'app-1', orgId: 'org-1', name: 'staging' })
  ).rejects.toBeTruthy()
})

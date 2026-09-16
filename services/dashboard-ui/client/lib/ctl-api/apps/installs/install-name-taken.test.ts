import { beforeEach, expect, mock, test } from 'bun:test'

type Page = { data: unknown[]; pagination: { hasNext: boolean } }

const pages: Page[] = []
const calls: Record<string, unknown>[] = []

const getAppInstalls = mock(async (args: Record<string, unknown>) => {
  calls.push(args)
  return pages.shift() ?? { data: [], pagination: { hasNext: false } }
})

mock.module('./get-app-installs', () => ({ getAppInstalls }))

const { installNameTaken } = await import('./install-name-taken')

beforeEach(() => {
  pages.length = 0
  calls.length = 0
})

test('reports a taken name on an exact match', async () => {
  pages.push({
    data: [{ name: 'staging-copy' }, { name: 'staging' }],
    pagination: { hasNext: false },
  })

  expect(
    await installNameTaken({ appId: 'app-1', orgId: 'org-1', name: 'staging' })
  ).toBe(true)
  expect(calls[0]).toMatchObject({ q: 'staging', appId: 'app-1' })
})

test('substring matches alone do not make a name taken', async () => {
  pages.push({
    data: [{ name: 'staging-copy' }, { name: 'pre-staging' }],
    pagination: { hasNext: false },
  })

  expect(
    await installNameTaken({ appId: 'app-1', orgId: 'org-1', name: 'staging' })
  ).toBe(false)
})

test('keeps paging until the exact match is found', async () => {
  pages.push({
    data: [{ name: 'staging-a' }, { name: 'staging-b' }],
    pagination: { hasNext: true },
  })
  pages.push({ data: [{ name: 'staging' }], pagination: { hasNext: false } })

  expect(
    await installNameTaken({ appId: 'app-1', orgId: 'org-1', name: 'staging' })
  ).toBe(true)
  expect(calls).toHaveLength(2)
  expect(calls[1]).toMatchObject({ offset: 2 })
})

test('an empty name is never taken and skips the lookup', async () => {
  expect(
    await installNameTaken({ appId: 'app-1', orgId: 'org-1', name: '   ' })
  ).toBe(false)
  expect(calls).toHaveLength(0)
})

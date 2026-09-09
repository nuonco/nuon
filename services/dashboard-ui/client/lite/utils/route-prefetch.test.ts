import { expect, test } from 'bun:test'
import { appsListQuery } from '../queries/apps'
import { installsListQuery } from '../queries/installs'
import { routePrefetchQueries } from './route-prefetch'

const keysFor = (href: string) =>
  routePrefetchQueries(href).map((query) => query.queryKey)

test('prefetches the default list query for a collection route', () => {
  expect(keysFor('/org-123/apps')).toEqual([
    appsListQuery({ orgId: 'org-123' }).queryKey,
  ])
  expect(keysFor('/org-123/installs')).toEqual([
    installsListQuery({ orgId: 'org-123' }).queryKey,
  ])
})

test('skips routes that have no prefetchable data', () => {
  expect(keysFor('/org-123')).toEqual([])
  expect(keysFor('/org-123/settings/webhooks')).toEqual([])
  expect(keysFor('/org-123/apps/app-1')).toEqual([])
})

test('skips links that would land on a different list state', () => {
  expect(keysFor('/org-123/apps?q=payments')).toEqual([])
  expect(keysFor('/org-123/installs?offset=20')).toEqual([])
  expect(keysFor('/org-123/installs#section')).toEqual([])
})

test('skips links that leave the app', () => {
  expect(keysFor('https://docs.nuon.co/installs')).toEqual([])
  expect(keysFor('mailto:support@nuon.co')).toEqual([])
})

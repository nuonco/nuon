import { afterEach, expect, test } from 'bun:test'
import { cleanup, render } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { useListQueryState } from '../hooks/use-list-query-state'
import { APPS_PAGE_SIZE, appsListQuery } from './apps'
import {
  INSTALL_FILTERS,
  INSTALLS_PAGE_SIZE,
  installsListQuery,
} from './installs'

afterEach(cleanup)

const mountedKey = (
  useKey: () => readonly unknown[],
  path: string
): readonly unknown[] => {
  let key: readonly unknown[] = []
  const Probe = () => {
    key = useKey()
    return null
  }

  render(
    <MemoryRouter initialEntries={[path]}>
      <Probe />
    </MemoryRouter>
  )

  return key
}

test('the apps prefetch key matches what the mounted list asks for', () => {
  const key = mountedKey(() => {
    const list = useListQueryState({ pageSize: APPS_PAGE_SIZE })
    return ['apps', 'org-123', ...list.queryKey]
  }, '/org-123/apps')

  expect(key).toEqual(appsListQuery({ orgId: 'org-123' }).queryKey)
})

test('the installs prefetch key matches what the mounted list asks for', () => {
  const key = mountedKey(() => {
    const list = useListQueryState({
      pageSize: INSTALLS_PAGE_SIZE,
      filters: INSTALL_FILTERS,
    })
    return ['installs', 'org-123', ...list.queryKey]
  }, '/org-123/installs')

  expect(key).toEqual(installsListQuery({ orgId: 'org-123' }).queryKey)
})

test('a filtered list asks for a different key than the prefetch warms', () => {
  const key = mountedKey(() => {
    const list = useListQueryState({
      pageSize: INSTALLS_PAGE_SIZE,
      filters: INSTALL_FILTERS,
    })
    return ['installs', 'org-123', ...list.queryKey]
  }, '/org-123/installs?labels=env:prod')

  expect(key).not.toEqual(installsListQuery({ orgId: 'org-123' }).queryKey)
})

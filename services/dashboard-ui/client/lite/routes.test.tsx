import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { matchRoutes, MemoryRouter } from 'react-router'
import { SubNav } from './components/molecules/SubNav'
import { orgNavigation } from './pages/OrgLayout'
import { settingsNavigation } from './pages/SettingsLayout'
import { liteRoutes } from './routes'

afterEach(cleanup)

const matchedIds = (path: string) =>
  matchRoutes(liteRoutes, path)?.map((match) => match.route.id)

test('matches focused and organization-scoped top-level routes', () => {
  expect(matchedIds('/onboarding')).toEqual([
    'root-layout',
    'focus-layout',
    'onboarding',
  ])
  expect(matchedIds('/org-123')).toEqual([
    'root-layout',
    'org-layout',
    'dashboard',
  ])
  expect(matchedIds('/org-123/apps')).toEqual([
    'root-layout',
    'org-layout',
    'apps',
  ])
  expect(matchedIds('/org-123/installs')).toEqual([
    'root-layout',
    'org-layout',
    'installs',
  ])
  expect(matchedIds('/org-123/teams')).toEqual([
    'root-layout',
    'org-layout',
    'teams',
  ])
})

test('matches every settings child from the playground route model', () => {
  expect(matchedIds('/org-123/settings')).toEqual([
    'root-layout',
    'org-layout',
    'settings-layout',
    'settings-connections',
  ])
  expect(matchedIds('/org-123/settings/webhooks')?.at(-1)).toBe(
    'settings-webhooks'
  )
  expect(matchedIds('/org-123/settings/triggers')?.at(-1)).toBe(
    'settings-triggers'
  )
  expect(matchedIds('/org-123/settings/api-tokens')?.at(-1)).toBe(
    'settings-api-tokens'
  )
  expect(matchedIds('/org-123/settings/service-accounts')?.at(-1)).toBe(
    'settings-service-accounts'
  )
  expect(matchedIds('/org-123/settings/oidc')?.at(-1)).toBe('settings-oidc')
})

test('leaves the bare root to the BFF and catches unknown org pages', () => {
  expect(matchRoutes(liteRoutes, '/')).toBeNull()
  expect(matchedIds('/org-123/unknown')?.at(-1)).toBe('org-not-found')
})

test('builds every shell destination from the active organization', () => {
  const navigation = orgNavigation('org-123')
  const destinations = [...navigation.primary, ...navigation.secondary]

  expect(destinations.find((item) => item.label === 'Dashboard')?.href).toBe(
    '/org-123'
  )
  expect(destinations.find((item) => item.label === 'Team')?.href).toBe(
    '/org-123/teams'
  )
  expect(destinations.find((item) => item.label === 'Settings')?.href).toBe(
    '/org-123/settings'
  )
})

test('marks the active settings section', () => {
  render(
    <MemoryRouter initialEntries={['/org-123/settings/webhooks']}>
      <SubNav items={settingsNavigation('org-123')} label="Settings sections" />
    </MemoryRouter>
  )

  expect(
    screen.getByRole('link', { name: 'Webhooks' }).getAttribute('aria-current')
  ).toBe('page')
  expect(
    screen
      .getByRole('link', { name: 'Connections' })
      .hasAttribute('aria-current')
  ).toBe(false)
})

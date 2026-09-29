import { describe, expect, test } from 'bun:test'
import {
  branchQuickNavHref,
  installQuickNavHref,
  withBreadcrumbQuickNav,
} from './breadcrumb-quick-nav'

describe('branchQuickNavHref', () => {
  const currentBasePath = '/org-1/apps/app-1/branches/branch-1'
  const targetBasePath = '/org-1/apps/app-1/branches/branch-2'

  test('preserves shared branch pages', () => {
    expect(
      branchQuickNavHref({
        pathname: `${currentBasePath}/rollout`,
        currentBasePath,
        targetBasePath,
      })
    ).toBe(`${targetBasePath}/rollout`)
  })

  test('falls back from resource details to their shared section', () => {
    expect(
      branchQuickNavHref({
        pathname: `${currentBasePath}/runs/run-1`,
        currentBasePath,
        targetBasePath,
      })
    ).toBe(`${targetBasePath}/runs`)
  })
})

describe('installQuickNavHref', () => {
  const currentBasePath = '/org-1/apps/app-1/installs/install-1'
  const targetBasePath = '/org-1/apps/app-2/installs/install-2'

  test('preserves shared new install pages', () => {
    expect(
      installQuickNavHref({
        pathname: `${currentBasePath}/configuration/inputs`,
        currentBasePath,
        targetBasePath,
      })
    ).toBe(`${targetBasePath}/configuration/inputs`)
  })

  test('falls back to the install root for resource details', () => {
    expect(
      installQuickNavHref({
        pathname: `${currentBasePath}/deployments/deploy-1`,
        currentBasePath,
        targetBasePath,
      })
    ).toBe(targetBasePath)
  })
})

describe('withBreadcrumbQuickNav', () => {
  test('attaches install quick nav only to the install name', () => {
    const installPath = '/org-1/apps/app-1/installs/install-1'
    const crumbs = withBreadcrumbQuickNav(
      [
        { path: '/org-1', text: 'Acme' },
        { path: '/org-1/apps/app-1', text: 'payments' },
        { path: installPath, text: 'production' },
        { path: installPath, text: 'Overview' },
        { path: `${installPath}/resources`, text: 'Resources' },
      ],
      {
        install: {
          enabled: true,
          orgId: 'org-1',
          appId: 'app-1',
          resourceId: 'install-1',
          name: 'production',
        },
      }
    )

    expect(
      crumbs.filter((crumb) => crumb.quickNav).map((crumb) => crumb.text)
    ).toEqual(['production'])
  })
})

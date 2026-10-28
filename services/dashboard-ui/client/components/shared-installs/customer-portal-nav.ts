import type { TNavItem } from '@/types'
import type { TInstallShareScope } from './types'

const BASE_LINKS: TNavItem[] = [
  {
    path: `/`,
    iconVariant: 'HouseSimpleIcon',
    text: 'Overview',
  },
  {
    path: `/deployments`,
    iconVariant: 'ArrowsClockwiseIcon',
    text: 'Deployments',
  },
]

const SCOPED_LINKS: Record<TInstallShareScope, TNavItem> = {
  approvals: {
    path: `/approvals`,
    iconVariant: 'ListChecksIcon',
    text: 'Approvals',
  },
  readme: {
    path: `/readme`,
    iconVariant: 'BookOpenTextIcon',
    text: 'README',
  },
  operations: {
    path: `/operations`,
    iconVariant: 'TerminalWindowIcon',
    text: 'Operations',
  },
  resources: {
    path: `/resources`,
    iconVariant: 'CardsIcon',
    text: 'Resources',
  },
}

export const customerPortalNav = (
  scopes: TInstallShareScope[] = []
): TNavItem[] => {
  const granted = new Set(scopes)
  return [
    ...BASE_LINKS,
    ...INSTALL_SHARE_SCOPE_ORDER.filter((scope) => granted.has(scope)).map(
      (scope) => SCOPED_LINKS[scope]
    ),
  ]
}

const INSTALL_SHARE_SCOPE_ORDER: TInstallShareScope[] = [
  'approvals',
  'readme',
  'resources',
  'operations',
]

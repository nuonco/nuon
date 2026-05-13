import type { TNavLink } from '@/types'

export const MAIN_LINKS: TNavLink[] = [
  {
    iconVariant: 'HouseIcon',
    path: `/`,
    text: 'Dashboard',
    shortcut: 'g d',
  },
  {
    iconVariant: 'AppWindowIcon',
    path: `/apps`,
    text: 'Apps',
    shortcut: 'g a',
  },
  {
    iconVariant: 'CubeIcon',
    path: `/installs`,
    text: 'Installs',
    shortcut: 'g i',
  },
]

export const SETTINGS_LINKS: TNavLink[] = [
  {
    iconVariant: 'UsersThreeIcon',
    path: `/team`,
    text: 'Team',
    shortcut: 'g t',
  },
  {
    iconVariant: 'HammerIcon',
    path: `/runner`,
    text: 'Build runner',
    shortcut: 'g r',
  },
  {
    iconVariant: 'WebhooksLogoIcon',
    path: `/webhooks`,
    text: 'Webhooks',
    shortcut: 'g w',
  },
]

export const SUPPORT_LINKS: TNavLink[] = [
  {
    iconVariant: 'BookOpenTextIcon',
    path: `https://docs.nuon.co/get-started/introduction`,
    text: 'Developer docs',
    isExternal: true,
  },
  // {
  //   iconVariant: 'ListBulletsIcon',
  //   path: `/releases`,
  //   text: 'Releases',
  // },
]

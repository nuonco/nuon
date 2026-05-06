import type { TNavLink } from '@/types'

export const MAIN_LINKS: TNavLink[] = [
  {
    iconVariant: 'House',
    path: `/`,
    text: 'Dashboard',
    shortcut: 'g d',
  },
  {
    iconVariant: 'AppWindow',
    path: `/apps`,
    text: 'Apps',
    shortcut: 'g a',
  },
  {
    iconVariant: 'Cube',
    path: `/installs`,
    text: 'Installs',
    shortcut: 'g i',
  },
]

export const SETTINGS_LINKS: TNavLink[] = [
  {
    iconVariant: 'UsersThree',
    path: `/team`,
    text: 'Team',
    shortcut: 'g t',
  },
  {
    iconVariant: 'Hammer',
    path: `/runner`,
    text: 'Build runner',
    shortcut: 'g r',
  },
  {
    iconVariant: 'WebhooksLogo',
    path: `/webhooks`,
    text: 'Webhooks',
    shortcut: 'g w',
  },
]

export const SUPPORT_LINKS: TNavLink[] = [
  {
    iconVariant: 'BookOpenText',
    path: `https://docs.nuon.co/get-started/introduction`,
    text: 'Developer docs',
    isExternal: true,
  },
  // {
  //   iconVariant: 'ListBullets',
  //   path: `/releases`,
  //   text: 'Releases',
  // },
]

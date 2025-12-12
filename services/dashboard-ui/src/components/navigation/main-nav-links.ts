import type { TNavLink } from '@/types'

export const MAIN_LINKS: TNavLink[] = [
  {
    iconVariant: 'House',
    path: `/`,
    text: 'Dashboard',
  },
  {
    iconVariant: 'AppWindow',
    path: `/apps`,
    text: 'Apps',
  },
  {
    iconVariant: 'Cube',
    path: `/installs`,
    text: 'Installs',
  },
  {
    iconVariant: 'SneakerMove',
    path: `/runner`,
    text: 'Build runner',
  },
]

export const SETTINGS_LINKS: TNavLink[] = [
  {
    iconVariant: 'UsersThree',
    path: `/team`,
    text: 'Team',
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

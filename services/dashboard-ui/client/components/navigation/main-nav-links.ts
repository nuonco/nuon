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
    text: 'Builds',
    shortcut: 'g r',
  },
  {
    iconVariant: 'LightningIcon',
    path: `/triggers`,
    text: 'Triggers',
    shortcut: 'g e',
  },
  {
    iconVariant: 'WebhooksLogoIcon',
    path: `/webhooks`,
    text: 'Webhooks',
    shortcut: 'g w',
  },
  {
    iconVariant: 'KeyIcon',
    path: `/api-tokens`,
    text: 'API tokens',
    shortcut: 'g k',
  },
  {
    iconVariant: 'RobotIcon',
    path: `/service-accounts`,
    text: 'Service accounts',
    shortcut: 'g v',
  },
]

export const SLACK_LINK: TNavLink = {
  iconVariant: 'SlackLogoIcon',
  path: `/slack`,
  text: 'Slack',
  shortcut: 'g s',
}

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

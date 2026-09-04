import { ComponentDocs } from '../__stories__/ComponentDocs'
import { DashboardNav } from './DashboardNav'

export default {
  title: 'lite/organisms/DashboardNav',
}

const ITEMS = [
  {
    href: '/',
    label: 'Dashboard',
    icon: 'HouseIcon' as const,
    shortcut: 'g d',
    end: true,
  },
  {
    href: '/apps',
    label: 'Apps',
    icon: 'AppWindowIcon' as const,
    shortcut: 'g a',
  },
  {
    href: '/installs',
    label: 'Installs',
    icon: 'CubeIcon' as const,
    shortcut: 'g i',
  },
]

export const Overview = () => (
  <ComponentDocs
    name="DashboardNav"
    tier="organism"
    summary="An accessible group of dashboard navigation links."
    use={[
      'Group related destinations in the dashboard sidebar.',
      'Use a separate group for secondary destinations anchored near the sidebar footer.',
    ]}
    avoid={[
      'Do not fetch route or organization configuration inside the group.',
      'Do not maintain a second shortcut configuration.',
    ]}
    rules={[
      'The group label is exposed to assistive technology without adding visible section chrome.',
      'Mobile navigation calls onNavigate after an internal selection.',
    ]}
    props={[
      {
        name: 'items',
        type: 'INavItem[]',
        description: 'Destinations rendered in the group.',
      },
      {
        name: 'label',
        type: 'string',
        description: 'Accessible group name.',
      },
      {
        name: 'collapsed',
        type: 'boolean',
        default: 'false',
        description: 'Uses icon-only navigation.',
      },
    ]}
  />
)

export const Expanded = () => (
  <div className="w-56 p-4">
    <DashboardNav items={ITEMS} label="Main" />
  </div>
)

export const Collapsed = () => (
  <div className="w-14 p-2">
    <DashboardNav items={ITEMS} label="Main" collapsed />
  </div>
)

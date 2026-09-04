import { ComponentDocs } from '../__stories__/ComponentDocs'
import { NavLink } from './NavLink'

export default {
  title: 'lite/molecules/NavLink',
}

export const Overview = () => (
  <ComponentDocs
    name="NavLink"
    tier="molecule"
    summary="A route-aware dashboard navigation link."
    use={[
      'Use for destinations in persistent application navigation.',
      'Provide the displayed keyboard chord from the same navigation configuration that handles it.',
    ]}
    avoid={[
      'Do not use for links inside page content.',
      'Do not hide labels outside the collapsed desktop sidebar.',
    ]}
    rules={[
      'Internal routes expose aria-current when active.',
      'Collapsed links expose their label and shortcut in a tooltip.',
      'External navigation opens in a new tab and never receives a route shortcut.',
    ]}
    props={[
      {
        name: 'href',
        type: 'string',
        description: 'Internal route or external destination.',
      },
      {
        name: 'label',
        type: 'string',
        description: 'Visible and accessible destination name.',
      },
      {
        name: 'icon',
        type: 'TIconVariant',
        description: 'Phosphor icon displayed before the label.',
      },
      {
        name: 'shortcut',
        type: 'string',
        description: 'Optional navigation chord.',
      },
      {
        name: 'collapsed',
        type: 'boolean',
        default: 'false',
        description: 'Renders an icon-only link with a tooltip.',
      },
    ]}
  />
)

export const Expanded = () => (
  <div className="w-56 p-8">
    <NavLink href="/" label="Dashboard" icon="HouseIcon" shortcut="g d" end />
  </div>
)

export const Collapsed = () => (
  <div className="w-14 p-2">
    <NavLink
      href="/"
      label="Dashboard"
      icon="HouseIcon"
      shortcut="g d"
      collapsed
      end
    />
  </div>
)

export const External = () => (
  <div className="w-56 p-8">
    <NavLink
      href="https://docs.nuon.co"
      label="Developer docs"
      icon="BookOpenTextIcon"
      external
    />
  </div>
)

export const LongLabel = () => (
  <div className="w-48 p-8">
    <NavLink
      href="/settings"
      label="Organization configuration and preferences"
      icon="GearIcon"
      shortcut="g s"
    />
  </div>
)

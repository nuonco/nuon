import { useState } from 'react'
import {
  Menu,
  MenuItem,
  MenuLabel,
  MenuSeparator,
  MenuSubmenu,
} from '../molecules/Menu'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Button } from './Button'
import { ID } from '../molecules/ID'
import { Badge } from './Badge'
import { Dropdown } from './Dropdown'
import { Icon } from './Icon'
import { Status } from './Status'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Dropdown',
}

export const Overview = () => (
  <ComponentDocs
    name="Dropdown"
    tier="atom"
    summary="A surface positioned against a trigger, opened on click."
    use={[
      'Hold a Menu of actions behind a Manage or kebab button.',
      'Hold a filter form, a date picker, or anything else that is not a list of commands.',
      'Give a control more room than it can have inline.',
    ]}
    avoid={[
      'Do not use it for a form field with a value. A select is a listbox and has its own keyboard contract.',
      'Do not use it to hide something the user has to find. A dropdown is for the second thing, not the only thing.',
      'Do not put a raw Dropdown inside a Menu. MenuSubmenu is the submenu, and it keeps the item roles and the keyboard right.',
    ]}
    rules={[
      'The trigger must pass the props it does not recognise to the element it renders. Button does. A trigger that swallows them never opens.',
      'A trigger tooltip goes on the Dropdown as triggerTooltip, not on the Button. The Dropdown is what knows to hide it while the menu is open.',
      'Set haspopup to match what is actually inside, so the trigger does not announce a menu it does not have.',
      'The surface has no padding of its own. Its contents own their spacing.',
    ]}
    props={[
      { name: 'trigger', type: 'ReactElement', description: 'The control that opens it. Cloned to add the ARIA and the handlers.' },
      { name: 'triggerTooltip', type: "Omit<ITooltip, 'children'>", description: 'Tooltip for the trigger. Forced closed while open.' },
      { name: 'open', type: 'boolean', description: 'Controlled mode.' },
      { name: 'defaultOpen', type: 'boolean', default: 'false', description: 'Uncontrolled initial state.' },
      { name: 'onOpenChange', type: '(open: boolean) => void', description: 'Fires on every open and close.' },
      { name: 'side', type: "'top' | 'bottom' | 'left' | 'right'", default: "'bottom'", description: 'Preferred side. Flips if it does not fit.' },
      { name: 'align', type: "'start' | 'center' | 'end'", default: "'start'", description: 'Alignment along the trigger edge.' },
      { name: 'haspopup', type: "'menu' | 'dialog' | 'listbox' | 'grid'", default: "'menu'", description: 'What the trigger announces it opens.' },
      { name: 'matchTriggerWidth', type: 'boolean', default: 'false', description: 'Surface is at least as wide as the trigger.' },
      { name: 'stretch', type: 'boolean', default: 'false', description: 'Trigger fills its container rather than hugging its content.' },
      { name: 'contentClassName', type: 'string', description: 'Classes for the floating surface.' },
    ]}
    sections={[
      {
        heading: 'Keyboard',
        body: 'Enter, Space and ArrowDown open and focus the first item; ArrowUp opens and focuses the last. Escape closes and puts focus back on the trigger. Tab closes and moves on. Clicking with a mouse opens without moving focus.',
      },
      {
        heading: 'Focus, and content that is not a menu',
        body: 'The Dropdown does not go looking for things to focus. Its contents register a first and last entry point on its context, which is how Menu hands over its items. Content that registers nothing keeps its own arrow keys, so a calendar grid behaves like a calendar grid.',
      },
      {
        heading: 'Nesting',
        body: 'A Dropdown inside another Dropdown registers its surface with the one above, so clicking into a submenu does not dismiss its parent, and choosing an item closes the whole stack. Escape closes only the level you are in. Because portalled content still bubbles its events through the React tree, both Menu and Dropdown stop the keys they have already handled, so one arrow press moves one level.',
      },
      {
        heading: 'Positioning',
        body: 'Rendered in a portal on the shared usePopover hook, so overflow and stacking contexts cannot clip it. It flips when the preferred side does not fit and gets a max height from the space it has, so a long menu scrolls rather than running off the screen. The surface is not built until it is first opened.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="p-20">
    <Dropdown
      trigger={
        <Button icon={<Icon variant="SlidersHorizontalIcon" size={16} />}>
          Manage
        </Button>
      }
    >
      <Menu>
        <MenuItem onSelect={() => {}}>Edit labels</MenuItem>
        <MenuItem onSelect={() => {}}>Audit history</MenuItem>
        <MenuSeparator />
        <MenuItem tone="danger" onSelect={() => {}}>
          Deprovision
        </MenuItem>
      </Menu>
    </Dropdown>
  </div>
)

export const KebabTrigger = () => (
  <div className="p-20">
    <Dropdown
      align="end"
      triggerTooltip={{ content: 'Manage install' }}
      trigger={
        <Button variant="ghost" iconOnly aria-label="Manage install">
          <Icon variant="DotsThreeIcon" size={16} />
        </Button>
      }
    >
      <Menu>
        <MenuItem onSelect={() => {}}>Reprovision</MenuItem>
        <MenuItem onSelect={() => {}}>Sync secrets</MenuItem>
      </Menu>
    </Dropdown>
  </div>
)

export const TriggerNudge = () => {
  const [nudge, setNudge] = useState(true)

  return (
    <div className="flex items-center gap-3 p-20">
      <Dropdown
        onOpenChange={(open) => open && setNudge(false)}
        triggerTooltip={{
          content: 'Everything you can do to this install lives here',
          open: nudge,
          disableHover: true,
          side: 'bottom',
        }}
        trigger={<Button variant="primary">Manage</Button>}
      >
        <Menu>
          <MenuItem onSelect={() => {}}>Reprovision</MenuItem>
          <MenuItem onSelect={() => {}}>Deprovision</MenuItem>
        </Menu>
      </Dropdown>
      <Button variant="ghost" onClick={() => setNudge(true)}>
        Show nudge
      </Button>
    </div>
  )
}

const SIDES = ['bottom', 'top', 'right', 'left'] as const
const ALIGNS = ['start', 'center', 'end'] as const

export const Placement = () => (
  <div className="flex flex-col gap-10 p-24">
    <Text as="p" variant="caption" color="tertiary" className="max-w-md">
      Side picks the edge it opens from, align picks where it sits along that
      edge. Both are preferences: a surface that will not fit flips to the
      opposite side and stays inside the viewport.
    </Text>
    <div className="grid grid-cols-3 gap-x-8 gap-y-14">
      {SIDES.map((side) =>
        ALIGNS.map((align) => (
          <div key={`${side}-${align}`} className="flex justify-center">
            <Dropdown side={side} align={align} trigger={<Button>{`${side} / ${align}`}</Button>}>
              <Menu className="min-w-40">
                <MenuItem onSelect={() => {}}>Reprovision</MenuItem>
                <MenuItem onSelect={() => {}}>Sync secrets</MenuItem>
                <MenuItem onSelect={() => {}}>Generate config</MenuItem>
              </Menu>
            </Dropdown>
          </div>
        ))
      )}
    </div>
  </div>
)

export const PlacementPlayground = () => {
  const [side, setSide] = useState<(typeof SIDES)[number]>('bottom')
  const [align, setAlign] = useState<(typeof ALIGNS)[number]>('start')

  return (
    <div className="flex h-[80vh] flex-col items-center justify-center gap-10">
      <div className="flex flex-col items-center gap-3">
        <div className="flex items-center gap-2">
          <Text variant="label" color="tertiary" className="w-12">
            Side
          </Text>
          {SIDES.map((option) => (
            <Button
              key={option}
              size="sm"
              variant={option === side ? 'primary' : 'ghost'}
              onClick={() => setSide(option)}
            >
              {option}
            </Button>
          ))}
        </div>
        <div className="flex items-center gap-2">
          <Text variant="label" color="tertiary" className="w-12">
            Align
          </Text>
          {ALIGNS.map((option) => (
            <Button
              key={option}
              size="sm"
              variant={option === align ? 'primary' : 'ghost'}
              onClick={() => setAlign(option)}
            >
              {option}
            </Button>
          ))}
        </div>
      </div>

      <Dropdown
        key={`${side}-${align}`}
        defaultOpen
        side={side}
        align={align}
        trigger={<Button variant="secondary">{`${side} / ${align}`}</Button>}
      >
        <Menu className="min-w-44">
          <MenuItem onSelect={() => {}}>Reprovision</MenuItem>
          <MenuItem onSelect={() => {}}>Sync secrets</MenuItem>
        </Menu>
      </Dropdown>
    </div>
  )
}

export const FlipsAtViewportEdge = () => (
  <div className="flex h-[85vh] flex-col justify-between p-4">
    <div className="flex justify-center">
      <Dropdown trigger={<Button>Pinned to the top edge</Button>}>
        <Menu className="min-w-44">
          <MenuItem onSelect={() => {}}>Opened below, as asked</MenuItem>
        </Menu>
      </Dropdown>
    </div>
    <div className="flex justify-center">
      <Dropdown trigger={<Button>Pinned to the bottom edge</Button>}>
        <Menu className="min-w-44">
          <MenuItem onSelect={() => {}}>Asked for below</MenuItem>
          <MenuItem onSelect={() => {}}>Flipped above</MenuItem>
        </Menu>
      </Dropdown>
    </div>
    <div className="flex justify-between">
      <Dropdown side="left" trigger={<Button>Asked for left</Button>}>
        <Menu className="min-w-44">
          <MenuItem onSelect={() => {}}>Flipped to the right</MenuItem>
        </Menu>
      </Dropdown>
      <Dropdown side="right" trigger={<Button>Asked for right</Button>}>
        <Menu className="min-w-44">
          <MenuItem onSelect={() => {}}>Flipped to the left</MenuItem>
        </Menu>
      </Dropdown>
    </div>
  </div>
)

export const LongMenuScrolls = () => (
  <div className="p-20">
    <Dropdown trigger={<Button>Every region</Button>}>
      <Menu>
        {Array.from({ length: 30 }, (_, index) => (
          <MenuItem key={index} onSelect={() => {}}>
            us-west-{index + 1}
          </MenuItem>
        ))}
      </Menu>
    </Dropdown>
  </div>
)

export const MatchTriggerWidth = () => (
  <div className="p-20">
    <Dropdown
      matchTriggerWidth
      trigger={
        <Button
          className="w-96 justify-between"
          icon={<Icon variant="CaretDownIcon" size={16} />}
        >
          A trigger considerably wider than its menu
        </Button>
      }
    >
      <Menu className="min-w-0">
        <MenuItem onSelect={() => {}}>Production</MenuItem>
        <MenuItem onSelect={() => {}}>Staging</MenuItem>
      </Menu>
    </Dropdown>
  </div>
)

export const NotAMenu = () => (
  <div className="p-20">
    <Dropdown
      haspopup="dialog"
      align="start"
      contentClassName="w-80"
      trigger={<Button variant="secondary">acme-production</Button>}
    >
      <div className="flex flex-col gap-3 p-3">
        <div className="flex items-center justify-between gap-3">
          <Text weight="medium">acme-production</Text>
          <Status status="active" />
        </div>
        <ID value="inst_01h9k2m4p6q8r0s2t4v6w8x0" truncate />
        <div className="flex flex-wrap gap-1">
          <Badge variant="code" labelKey="env" labelValue="production" />
          <Badge variant="code" labelKey="region" labelValue="us-west-2" />
        </div>
        <Text variant="caption" color="secondary">
          Deployed 4 minutes ago by a service account.
        </Text>
        <div className="flex justify-end gap-2">
          <Button size="sm" variant="ghost">
            View logs
          </Button>
          <Button size="sm">Open install</Button>
        </div>
      </div>
    </Dropdown>
    <Text as="p" variant="caption" color="tertiary" className="mt-6 max-w-md">
      Not every dropdown holds a list of commands. This one announces itself as
      a dialog, keeps its own keyboard, and owns the padding inside the surface.
      A date picker will sit here the same way.
    </Text>
  </div>
)

export const InGroupedActions = () => (
  <div className="p-20">
    <Dropdown
      align="end"
      trigger={
        <Button
          variant="primary"
          icon={<Icon variant="SlidersHorizontalIcon" size={16} />}
        >
          Manage
        </Button>
      }
    >
      <Menu className="min-w-56">
        <MenuLabel>Settings</MenuLabel>
        <MenuItem onSelect={() => {}}>Edit labels</MenuItem>
        <MenuItem onSelect={() => {}}>Audit history</MenuItem>
        <MenuItem href="/docs/installs">Install documentation</MenuItem>
        <MenuSeparator />
        <MenuLabel>Controls</MenuLabel>
        <MenuItem onSelect={() => {}}>Reprovision</MenuItem>
        <MenuItem disabled onSelect={() => {}}>
          Sync secrets
        </MenuItem>
        <MenuSeparator />
        <MenuLabel>Danger</MenuLabel>
        <MenuItem
          tone="danger"
          icon={<Icon variant="TrashIcon" size={16} />}
          onSelect={() => {}}
        >
          Deprovision
        </MenuItem>
      </Menu>
    </Dropdown>
  </div>
)

export const NestedMenus = () => (
  <div className="p-20">
    <Dropdown
      trigger={
        <Button icon={<Icon variant="SlidersHorizontalIcon" size={16} />}>
          Manage
        </Button>
      }
    >
      <Menu className="min-w-56">
        <MenuLabel>Settings</MenuLabel>
        <MenuItem onSelect={() => {}}>Edit labels</MenuItem>
        <MenuSubmenu label="Notifications">
          <Menu className="min-w-48">
            <MenuItem onSelect={() => {}}>Slack channel</MenuItem>
            <MenuItem onSelect={() => {}}>Webhook</MenuItem>
            <MenuSubmenu label="Advanced">
              <Menu className="min-w-44">
                <MenuItem onSelect={() => {}}>Retry policy</MenuItem>
                <MenuItem onSelect={() => {}}>Delivery log</MenuItem>
              </Menu>
            </MenuSubmenu>
          </Menu>
        </MenuSubmenu>
        <MenuSeparator />
        <MenuLabel>Controls</MenuLabel>
        <MenuSubmenu
          icon={<Icon variant="ArrowClockwiseIcon" size={16} />}
          label="Reprovision"
        >
          <Menu className="min-w-48">
            <MenuItem onSelect={() => {}}>Install</MenuItem>
            <MenuItem onSelect={() => {}}>Sandbox</MenuItem>
            <MenuItem onSelect={() => {}}>Stack</MenuItem>
          </Menu>
        </MenuSubmenu>
        <MenuItem onSelect={() => {}}>Sync secrets</MenuItem>
        <MenuSeparator />
        <MenuItem
          tone="danger"
          icon={<Icon variant="TrashIcon" size={16} />}
          onSelect={() => {}}
        >
          Deprovision
        </MenuItem>
      </Menu>
    </Dropdown>
    <Text as="p" variant="caption" color="tertiary" className="mt-6 max-w-md">
      Right arrow opens a submenu, left arrow closes it, and up and down stay in
      the level you are in. Choosing an action closes every level at once.
    </Text>
  </div>
)

import { useState } from 'react'
import { Button } from '../atoms/Button'
import { Dropdown } from '../atoms/Dropdown'
import { Icon } from '../atoms/Icon'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Menu, MenuItem, MenuLabel, MenuSeparator, MenuSubmenu } from './Menu'

export default {
  title: 'lite/molecules/Menu',
}

export const Overview = () => (
  <ComponentDocs
    name="Menu"
    tier="molecule"
    summary="A list of commands, usually inside a Dropdown."
    use={[
      'Group the actions available on a resource behind one trigger.',
      'Offer a set of choices where picking one is the whole interaction.',
      'Filter a list, with items that stay open as you tick them.',
    ]}
    avoid={[
      'Do not use it as a form field. A select is a listbox with a value, and it belongs with the form fields.',
      'Do not put a Button or a Link inside it. MenuItem is the item, and it is the only item.',
      'Do not build a submenu by hand. MenuSubmenu is a menu item that opens another Menu beside it.',
    ]}
    rules={[
      'Every child is a MenuItem, a MenuLabel or a MenuSeparator. The Menu does not restyle anything it is given.',
      'A MenuItem closes the menu when it is chosen. Pass closeOnSelect={false} for the multi-select case, which is the only reason to keep it open.',
      'Pass selected to make an item checkable. It renders a check and announces itself as one.',
      'Keep item labels to a verb and its object, so typeahead lands where the user expects.',
    ]}
    props={[
      { name: 'icon', type: 'ReactNode', description: 'MenuItem: leading icon.' },
      { name: 'href', type: 'string', description: 'MenuItem: renders a link. Internal or external is inferred.' },
      { name: 'selected', type: 'boolean', description: 'MenuItem: checkable item. Renders a check.' },
      { name: 'disabled', type: 'boolean', default: 'false', description: 'MenuItem: unavailable, and skipped by arrow keys.' },
      { name: 'tone', type: "'default' | 'danger'", default: "'default'", description: 'MenuItem: danger for destructive actions.' },
      { name: 'closeOnSelect', type: 'boolean', default: 'true', description: 'MenuItem: whether choosing it closes the Dropdown.' },
      { name: 'onSelect', type: '() => void', description: 'MenuItem: fires on click, Enter and Space.' },
      { name: 'label', type: 'ReactNode', description: 'MenuSubmenu: the item text. Its children are the Menu it opens.' },
    ]}
    sections={[
      {
        heading: 'Keyboard',
        body: 'The Menu owns the keyboard because it owns the role. Up and Down move and wrap, Home and End jump to the ends, and typing moves to the next item starting with what you typed. Only one item is ever in the tab order, so tabbing leaves the menu rather than walking through it. Disabled items are skipped.',
      },
      {
        heading: 'Submenus',
        body: 'MenuSubmenu is a menu item that opens another Menu to its right. Right arrow opens it, left arrow closes it, and up and down stay in the level you are in. Choosing an action anywhere in the stack closes all of it. Keep them to one level where you can, because a submenu hides its contents from anyone scanning the menu.',
      },
      {
        heading: 'Outside a Dropdown',
        body: 'A Menu works on its own, for a sidebar of commands or a story like this one. It looks for a Dropdown to close and finds nothing, so choosing an item just runs its action.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="w-fit rounded-lg border border-divider bg-popover-bg p-0">
    <Menu>
      <MenuItem onSelect={() => {}}>Edit labels</MenuItem>
      <MenuItem onSelect={() => {}}>Audit history</MenuItem>
      <MenuItem onSelect={() => {}}>Generate config</MenuItem>
    </Menu>
  </div>
)

export const WithIcons = () => (
  <div className="w-fit rounded-lg border border-divider bg-popover-bg">
    <Menu>
      <MenuItem icon={<Icon variant="ArrowClockwiseIcon" size={16} />} onSelect={() => {}}>
        Reprovision
      </MenuItem>
      <MenuItem icon={<Icon variant="PlusIcon" size={16} />} onSelect={() => {}}>
        Add component
      </MenuItem>
      <MenuItem
        tone="danger"
        icon={<Icon variant="TrashIcon" size={16} />}
        onSelect={() => {}}
      >
        Deprovision
      </MenuItem>
    </Menu>
  </div>
)

export const Grouped = () => (
  <div className="w-fit rounded-lg border border-divider bg-popover-bg">
    <Menu className="min-w-56">
      <MenuLabel>Settings</MenuLabel>
      <MenuItem onSelect={() => {}}>Edit labels</MenuItem>
      <MenuItem onSelect={() => {}}>Audit history</MenuItem>
      <MenuSeparator />
      <MenuLabel>Controls</MenuLabel>
      <MenuItem onSelect={() => {}}>Reprovision</MenuItem>
      <MenuItem disabled onSelect={() => {}}>
        Sync secrets
      </MenuItem>
      <MenuSeparator />
      <MenuLabel>Danger</MenuLabel>
      <MenuItem tone="danger" onSelect={() => {}}>
        Deprovision
      </MenuItem>
    </Menu>
  </div>
)

export const Links = () => (
  <div className="w-fit rounded-lg border border-divider bg-popover-bg">
    <Menu>
      <MenuItem href="/installs">All installs</MenuItem>
      <MenuItem href="/settings/team">Team settings</MenuItem>
      <MenuItem href="https://docs.nuon.co">Documentation</MenuItem>
    </Menu>
  </div>
)

export const Switcher = () => {
  const [branch, setBranch] = useState('main')
  const branches = ['main', 'staging', 'feat/new-runner']

  return (
    <div className="p-20">
      <Dropdown
        trigger={
          <Button icon={<Icon variant="CaretDownIcon" size={16} />}>
            {branch}
          </Button>
        }
      >
        <Menu>
          {branches.map((name) => (
            <MenuItem
              key={name}
              selected={name === branch}
              onSelect={() => setBranch(name)}
            >
              {name}
            </MenuItem>
          ))}
        </Menu>
      </Dropdown>
    </div>
  )
}

export const MultiSelectFilter = () => {
  const [selected, setSelected] = useState<string[]>(['env:production'])
  const labels = ['env:production', 'env:staging', 'tier:web', 'tier:worker']

  const toggle = (label: string) =>
    setSelected((current) =>
      current.includes(label)
        ? current.filter((item) => item !== label)
        : [...current, label]
    )

  return (
    <div className="p-20">
      <Dropdown
        trigger={
          <Button icon={<Icon variant="FunnelIcon" size={16} />}>
            Labels{selected.length ? ` (${selected.length})` : ''}
          </Button>
        }
      >
        <Menu className="min-w-56">
          <MenuLabel>Filter by label</MenuLabel>
          {labels.map((label) => (
            <MenuItem
              key={label}
              closeOnSelect={false}
              selected={selected.includes(label)}
              onSelect={() => toggle(label)}
            >
              {label}
            </MenuItem>
          ))}
          <MenuSeparator />
          <MenuItem onSelect={() => setSelected([])}>Clear filters</MenuItem>
        </Menu>
      </Dropdown>
    </div>
  )
}

export const Submenus = () => (
  <div className="p-20">
    <Dropdown trigger={<Button>Manage</Button>}>
      <Menu className="min-w-56">
        <MenuItem onSelect={() => {}}>Edit labels</MenuItem>
        <MenuSubmenu label="Reprovision">
          <Menu className="min-w-44">
            <MenuItem onSelect={() => {}}>Install</MenuItem>
            <MenuItem onSelect={() => {}}>Sandbox</MenuItem>
            <MenuItem onSelect={() => {}}>Stack</MenuItem>
          </Menu>
        </MenuSubmenu>
        <MenuSeparator />
        <MenuItem tone="danger" onSelect={() => {}}>
          Deprovision
        </MenuItem>
      </Menu>
    </Dropdown>
  </div>
)

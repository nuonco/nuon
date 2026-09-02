import { useState } from 'react'
import { Badge } from '../atoms/Badge'
import { Button } from '../atoms/Button'
import { Dropdown } from '../atoms/Dropdown'
import { Icon } from '../atoms/Icon'
import { Status } from '../atoms/Status'
import { Text } from '../atoms/Text'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Disclosure } from './Disclosure'
import { DisclosureGroup, ExpandAllButton } from './DisclosureGroup'
import { Menu, MenuItem } from './Menu'

export default {
  title: 'lite/molecules/Disclosure',
}

const Body = ({ children }: { children?: string }) => (
  <div className="px-2 pb-3 pt-1">
    <Text variant="caption" color="secondary">
      {children ??
        'Body content. It is not in the DOM until the section is opened, and it leaves again once the closing transition finishes.'}
    </Text>
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="Disclosure"
    tier="molecule"
    summary="A titled section that shows and hides its content."
    use={[
      'Break a long page into sections a reader can skip.',
      'Hold something expensive, such as a diff, that should not render until it is asked for.',
      'Group several sections so they can all be opened or closed at once.',
    ]}
    avoid={[
      'Do not use it to hide something the reader has to find. A section header has to say what is inside.',
      'Do not put the primary action of a page inside one.',
      'Do not nest them more than one deep. Two levels of caret is a tree, and a tree is a different component.',
    ]}
    rules={[
      'Controlled or uncontrolled, never both. Pass open to own it, defaultOpen to let it own itself.',
      'Anything interactive in the header goes in actions, never in title. The header itself is a button and cannot contain another one.',
      'Content unmounts when closed, so it must not be the only home for state you need to keep.',
      'Inside a DisclosureGroup the group sets the starting state; it does not keep members in step afterwards.',
    ]}
    props={[
      { name: 'title', type: 'ReactNode', description: 'The heading. Required.' },
      { name: 'description', type: 'ReactNode', description: 'Secondary line under the title.' },
      { name: 'status', type: 'ReactNode', description: 'Trailing slot inside the button, for a Badge or Status.' },
      { name: 'actions', type: 'ReactNode', description: 'Controls rendered beside the button, outside it.' },
      { name: 'icon', type: 'ReactNode', description: 'Leading icon after the caret.' },
      { name: 'open', type: 'boolean', description: 'Controlled mode.' },
      { name: 'defaultOpen', type: 'boolean', default: 'false', description: 'Uncontrolled initial state, or the group default.' },
      { name: 'onOpenChange', type: '(open: boolean) => void', description: 'Fires on every open and close, including from a group.' },
    ]}
    sections={[
      {
        heading: 'Why content unmounts',
        body: 'A page of sections is only cheap if a closed one costs nothing, which matters most when the body is a syntax-highlighted diff. Content mounts on first open and unmounts once the closing transition ends, so the animation still runs and nothing renders while closed. The cost is that a section forgets its own state — a scrolled body returns to the top.',
      },
      {
        heading: 'The animation',
        body: 'Height animates through grid template rows, from 0fr to 1fr, so it follows the real height of the content with no maximum to guess at and nothing to measure. Reduced motion collapses the duration rather than removing the transition, so the component still knows when the close has finished.',
      },
      {
        heading: 'Groups',
        body: 'A DisclosureGroup gives every member inside it a starting state and an expand-all control, without any of it being passed through props. Members register themselves, so the group knows how many are open and the control can flip between expand and collapse. Opening one member does not close the others.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-xl p-8">
    <Disclosure title="Deployment details">
      <Body />
    </Disclosure>
  </div>
)

export const OpenByDefault = () => (
  <div className="max-w-xl p-8">
    <Disclosure defaultOpen title="Deployment details" description="apps/api">
      <Body />
    </Disclosure>
  </div>
)

export const WithStatusAndActions = () => (
  <div className="max-w-xl p-8">
    <Disclosure
      title="acme-production"
      description="Deployment · apps/api"
      status={<Badge tone="warn">update</Badge>}
      actions={
        <Dropdown
          align="end"
          triggerTooltip={{ content: 'Manage section' }}
          trigger={
            <Button variant="ghost" iconOnly aria-label="Manage section">
              <Icon variant="DotsThreeIcon" size={16} />
            </Button>
          }
        >
          <Menu>
            <MenuItem onSelect={() => {}}>Copy manifest</MenuItem>
            <MenuItem onSelect={() => {}}>View resource</MenuItem>
          </Menu>
        </Dropdown>
      }
    >
      <Body>
        The manage menu sits beside the header button rather than inside it, so
        there is never a button within a button.
      </Body>
    </Disclosure>
  </div>
)

export const Controlled = () => {
  const [open, setOpen] = useState(false)

  return (
    <div className="flex max-w-xl flex-col gap-4 p-8">
      <div className="flex items-center gap-2">
        <Button size="sm" onClick={() => setOpen((current) => !current)}>
          Toggle from outside
        </Button>
        <Text variant="caption" color="tertiary">
          {open ? 'Open' : 'Closed'}
        </Text>
      </div>
      <Disclosure
        open={open}
        onOpenChange={setOpen}
        title="Owned by the parent"
        description="Clicking the header asks the parent, it does not decide"
      >
        <Body />
      </Disclosure>
    </div>
  )
}

export const TallContent = () => (
  <div className="max-w-xl p-8">
    <Disclosure
      title="Two hundred lines"
      description="Taller than any ceiling the old component guessed at"
    >
      <div className="flex flex-col gap-0.5 px-2 pb-3">
        {Array.from({ length: 200 }, (_, index) => (
          <Text key={index} variant="caption" family="mono" color="tertiary">
            line {index + 1}
          </Text>
        ))}
      </div>
    </Disclosure>
  </div>
)

const SECTIONS = [
  { name: 'aws_eks_cluster.this', detail: 'Cluster · us-west-2', status: 'active' },
  { name: 'aws_iam_role.runner', detail: 'Role · us-west-2', status: 'pending' },
  { name: 'helm_release.ingress', detail: 'Release · apps', status: 'failed' },
]

export const Group = () => (
  <div className="max-w-xl p-8">
    <DisclosureGroup className="gap-1">
      <div className="flex items-center justify-between pb-1">
        <Text variant="label" color="tertiary">
          Resource changes
        </Text>
        <ExpandAllButton />
      </div>
      {SECTIONS.map((section) => (
        <Disclosure
          key={section.name}
          title={section.name}
          description={section.detail}
          status={<Status status={section.status} />}
        >
          <Body />
        </Disclosure>
      ))}
    </DisclosureGroup>
  </div>
)

export const GroupOpenByDefault = () => (
  <div className="max-w-xl p-8">
    <DisclosureGroup defaultOpen className="gap-1">
      <div className="flex items-center justify-between pb-1">
        <Text variant="label" color="tertiary">
          Resource changes
        </Text>
        <ExpandAllButton />
      </div>
      {SECTIONS.map((section) => (
        <Disclosure
          key={section.name}
          title={section.name}
          description={section.detail}
          status={<Status status={section.status} />}
        >
          <Body>
            Collapsing one section leaves the others alone. The group sets where
            they start, not where they stay.
          </Body>
        </Disclosure>
      ))}
    </DisclosureGroup>
  </div>
)

import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Badge } from './Badge'
import { Button } from './Button'
import { Icon } from './Icon'
import { Text } from './Text'
import { Tooltip } from './Tooltip'

export default {
  title: 'lite/atoms/Tooltip',
}

export const Overview = () => (
  <ComponentDocs
    name="Tooltip"
    tier="atom"
    summary="A hover- and focus-triggered label for a control."
    use={[
      'Name an icon-only control.',
      'Explain why a control is disabled, though on a button you should use its own tooltip prop.',
      'Show the full value of something truncated.',
      'Drive a nudge from app state by passing open instead of relying on hover.',
    ]}
    avoid={[
      'Do not put content the user needs to act on in a tooltip. Tooltips vanish on blur and are unreachable on touch.',
      'Do not hide anything essential that is not also available another way.',
      'Do not wrap a button by hand, because the button takes a tooltip prop.',
    ]}
    rules={[
      'String content is wrapped and sized for you. Pass a node only for genuinely rich content.',
      'The side is a preference. It flips to the opposite side when there is not room.',
      'Pair open with disableHover for a nudge, so hover cannot fight the app state.',
    ]}
    props={[
      { name: 'content', type: 'ReactNode', description: 'Tooltip body. A string is wrapped in Text.' },
      { name: 'side', type: "'top' | 'bottom' | 'left' | 'right'", default: "'top'", description: 'Preferred side. Flips if it does not fit.' },
      { name: 'open', type: 'boolean', description: 'Controlled mode, for nudges driven by app state.' },
      { name: 'defaultOpen', type: 'boolean', default: 'false', description: 'Uncontrolled initial state.' },
      { name: 'onOpenChange', type: '(open: boolean) => void', description: 'Fires on hover and focus changes.' },
      { name: 'disableHover', type: 'boolean', default: 'false', description: 'Ignores hover and focus. Pair with open.' },
      { name: 'contentClassName', type: 'string', description: 'Classes for the floating surface.' },
    ]}
    sections={[
      {
        heading: 'Positioning',
        body: 'Rendered in a portal, so overflow and stacking contexts cannot clip it. It flips when the preferred side does not fit, stays inside the viewport, and keeps its arrow pointing at the trigger. Placement is shared with other floating surfaces via the usePopover hook.',
      },
    ]}
  />
)

export const Sides = () => (
  <div className="grid grid-cols-2 gap-8 p-20">
    {(['top', 'bottom', 'left', 'right'] as const).map((side) => (
      <div key={side} className="flex justify-center">
        <Tooltip side={side} content={`Positioned ${side}`}>
          <Button variant="secondary">{side}</Button>
        </Tooltip>
      </div>
    ))}
  </div>
)

export const OnIconButton = () => (
  <div className="flex items-center gap-3 p-20">
    <Tooltip content="Close">
      <Button variant="ghost" iconOnly aria-label="Close">
        <Icon variant="XIcon" size={16} />
      </Button>
    </Tooltip>
    <Tooltip content="Refresh the install list">
      <Button variant="secondary" iconOnly aria-label="Refresh">
        <Icon variant="ArrowClockwiseIcon" size={16} />
      </Button>
    </Tooltip>
  </div>
)

export const DisabledReason = () => (
  <div className="p-20">
    <Tooltip content="Sync the app config first">
      <Button variant="primary" disabled>
        Create install
      </Button>
    </Tooltip>
  </div>
)

export const OnTruncatedLabel = () => (
  <div className="max-w-xs p-20">
    <Tooltip content="a-very-long-label-key-that-keeps-going:and-an-even-longer-value">
      <Badge
        variant="code"
        labelKey="a-very-long-label-key-that-keeps-going"
        labelValue="and-an-even-longer-value"
      />
    </Tooltip>
  </div>
)

export const FlipsAtViewportEdge = () => (
  <div className="flex h-[90vh] flex-col justify-between p-4">
    <Tooltip side="top" content="Asked for top, flipped to bottom">
      <Button variant="secondary">Pinned to the top edge</Button>
    </Tooltip>
    <Tooltip side="bottom" content="Asked for bottom, flipped to top">
      <Button variant="secondary">Pinned to the bottom edge</Button>
    </Tooltip>
  </div>
)

export const Nudge = () => {
  const [open, setOpen] = useState(true)

  return (
    <div className="flex items-center gap-3 p-20">
      <Tooltip
        open={open}
        disableHover
        side="bottom"
        content="Trigger a run to deploy this branch"
      >
        <Button variant="primary" onClick={() => setOpen(false)}>
          Trigger run
        </Button>
      </Tooltip>
      <Button variant="ghost" onClick={() => setOpen(true)}>
        Show nudge
      </Button>
    </div>
  )
}

export const RichContent = () => (
  <div className="p-20">
    <Tooltip
      side="right"
      content={
        <div className="flex flex-col gap-1">
          <Text variant="caption" weight="medium">
            acme-production
          </Text>
          <Text variant="caption">Last deployed 4 minutes ago</Text>
        </div>
      }
    >
      <Button variant="secondary">Hover for detail</Button>
    </Tooltip>
  </div>
)

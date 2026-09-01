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
    summary="A hover- and focus-triggered label for a control. Positioned by usePopover, which owns measuring, flipping and clamping for every floating surface."
    use={[
      'Naming an icon-only control.',
      'Explaining why a control is disabled.',
      'Showing the full value of something truncated.',
      'A nudge — pass open to drive it from app state instead of hover.',
    ]}
    avoid={[
      'Content the user needs to act on. Tooltips vanish on blur and are unreachable on touch; use a Popover or inline text.',
      'Anything essential that is not also available another way.',
    ]}
    rules={[
      'Opens on focus as well as hover, so it is reachable by keyboard.',
      'The trigger gets aria-describedby pointing at the tooltip, so it is announced rather than being decoration.',
      'A string content auto-wraps in Text at caption size — callers do not size it themselves.',
      'Rendered in a portal on document.body, so overflow and stacking contexts cannot clip it.',
    ]}
    props={[
      { name: 'content', type: 'ReactNode', description: 'Tooltip body. A string is wrapped in Text.' },
      { name: 'side', type: "'top' | 'bottom' | 'left' | 'right'", default: "'top'", description: 'Preferred side. Flips to the opposite side if it does not fit.' },
      { name: 'open', type: 'boolean', description: 'Controlled mode, for nudges driven by app state.' },
      { name: 'defaultOpen', type: 'boolean', default: 'false', description: 'Uncontrolled initial state.' },
      { name: 'onOpenChange', type: '(open: boolean) => void', description: 'Fires on hover and focus changes.' },
      { name: 'disableHover', type: 'boolean', default: 'false', description: 'Ignores hover and focus. Pair with open for a pure nudge.' },
    ]}
    sections={[
      {
        heading: 'usePopover',
        body: 'The hook returns triggerRef, contentRef, the resolved side and a style object. It measures both elements, flips to the opposite side only when the preferred one does not fit and the opposite one does, clamps into the viewport with an 8px margin, and computes an --arrow offset that tracks the trigger centre but stays 12px from the tip edges so the arrow never falls off a corner. Repositioning is driven by a ResizeObserver on both elements plus resize and capture-phase scroll listeners — the old dashboard chased a moving trigger with a 30-frame requestAnimationFrame loop instead.',
      },
      {
        heading: 'Why a shared primitive',
        body: 'Tooltip and Dropdown in the dashboard each hand-rolled measure, flip, clamp and portal, with different names for the same axes — top/bottom/left/right in one, above/below/beside/overlay in the other. Dropdown will consume this hook too and keeps only what is genuinely its own: click to open, outside-dismiss and focus management.',
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

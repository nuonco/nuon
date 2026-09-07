import { Fragment, useState } from 'react'
import { Icon } from './Icon'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Button, type TButtonVariant } from './Button'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Button',
}

const VARIANTS: Array<{ variant: TButtonVariant; label: string; usage: string }> = [
  {
    variant: 'primary',
    label: 'Create install',
    usage: 'The one action the page exists for. At most one per page.',
  },
  {
    variant: 'secondary',
    label: 'Edit config',
    usage: 'The default. Every ordinary action.',
  },
  {
    variant: 'ghost',
    label: 'Cancel',
    usage: 'Lowest emphasis. Dismissals, toolbar actions, row actions.',
  },
  {
    variant: 'danger',
    label: 'Delete install',
    usage: 'Entering a destructive flow. Never the confirm-and-go button.',
  },
]

export const Overview = () => (
  <ComponentDocs
    name="Button"
    tier="atom"
    summary="Triggers an action. Four variants chosen by emphasis."
    use={[
      'Primary is the action the page exists for.',
      'Secondary is the default, for every ordinary action.',
      'Ghost is the lowest emphasis, for dismissals, toolbar actions and row actions.',
      'Danger marks the entry into a destructive flow.',
    ]}
    avoid={[
      'Do not use a button for navigation. Anything that changes the URL is a link, even when it looks like a button.',
      'Do not use danger for the confirm button inside a destructive flow. That modal has already said what will happen, so its confirm is a primary.',
      'Do not use the small size for a text button. It exists for icon-only affordances that sit inline with text.',
    ]}
    rules={[
      'At most one primary per page.',
      'Icon-only is a shape rather than a variant. Any variant can be icon-only, and it always needs an aria-label.',
      'Give every disabled button a tooltip explaining why, unless the reason is obvious. Pass the tooltip prop rather than wrapping the button by hand.',
      'Loading disables the button and shows a spinner. Use it for the gap between the click and the response.',
    ]}
    props={[
      { name: 'variant', type: "'primary' | 'secondary' | 'ghost' | 'danger'", default: "'secondary'", description: 'Emphasis level.' },
      { name: 'size', type: "'md' | 'sm'", default: "'md'", description: 'The small size is for icon-only affordances inline with text.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'Shows a spinner before the label, disables the button and sets aria-busy.' },
      { name: 'icon', type: 'ReactNode', description: 'Leading icon. Replaced by the spinner while loading.' },
      { name: 'iconOnly', type: 'boolean', default: 'false', description: 'Square button. Icon goes in children; aria-label is required.' },
      { name: 'tooltip', type: 'ReactNode', description: 'Wraps the button in a Tooltip. The way to explain a disabled state.' },
      { name: 'tooltipSide', type: "'top' | 'bottom' | 'left' | 'right'", default: "'top'", description: 'Preferred tooltip side.' },
      { name: 'disabled', type: 'boolean', description: 'Blocks activation and suppresses hover, while staying focusable so its tooltip is reachable.' },
    ]}
    sections={[
      {
        heading: 'States',
        body: 'Hover, active, disabled and focus-visible are defined for every variant, and hover is suppressed while disabled. All variants are the same height, so they line up in a row.',
      },
    ]}
  />
)

export const Variants = () => (
  <div className="flex flex-col gap-6 p-8">
    {VARIANTS.map(({ variant, label, usage }) => (
      <div key={variant} className="flex flex-col gap-2">
        <Text variant="label" color="tertiary" family="mono">
          {variant}
        </Text>
        <div>
          <Button variant={variant}>{label}</Button>
        </div>
        <Text variant="caption" color="tertiary">
          {usage}
        </Text>
      </div>
    ))}
  </div>
)

export const States = () => (
  <div className="flex flex-col gap-6 p-8">
    <Text variant="caption" color="tertiary">
      Hover, focus and active are live — tab to a button to see the focus ring.
      Disabled and loading are props.
    </Text>
    <div className="grid grid-cols-[auto_repeat(3,minmax(0,1fr))] items-center gap-4">
      <span />
      {['default', 'disabled', 'loading'].map((state) => (
        <Text key={state} variant="label" color="tertiary" family="mono">
          {state}
        </Text>
      ))}
      {VARIANTS.map(({ variant, label }) => (
        <Fragment key={variant}>
          <Text variant="label" color="tertiary" family="mono">
            {variant}
          </Text>
          <div>
            <Button variant={variant}>{label}</Button>
          </div>
          <div>
            <Button variant={variant} disabled>
              {label}
            </Button>
          </div>
          <div>
            <Button variant={variant} loading>
              {label}
            </Button>
          </div>
        </Fragment>
      ))}
    </div>
  </div>
)

export const WithIcon = () => (
  <div className="flex flex-wrap items-center gap-3 p-8">
    <Button variant="primary" icon={<Icon variant="PlusIcon" size={16} />}>
      Create install
    </Button>
    <Button icon={<Icon variant="ArrowClockwiseIcon" size={16} />}>Retry</Button>
    <Button variant="danger" icon={<Icon variant="TrashIcon" size={16} />}>
      Delete
    </Button>
  </div>
)

export const IconOnly = () => (
  <div className="flex items-center gap-3 p-8">
    <Button variant="ghost" iconOnly aria-label="Close">
      <Icon variant="XIcon" size={16} />
    </Button>
    <Button variant="secondary" iconOnly aria-label="Refresh">
      <Icon variant="ArrowClockwiseIcon" size={16} />
    </Button>
    <Text variant="caption" color="tertiary">
      iconOnly buttons require an aria-label. The icon goes in children.
    </Text>
  </div>
)

export const LoadingDoesNotJump = () => {
  const [loading, setLoading] = useState(false)

  return (
    <div className="flex flex-col items-start gap-6 p-8">
      <Text variant="caption" color="tertiary">
        The spinner column animates from 0fr to 1fr, so the button grows into the
        spinner rather than snapping. The label never moves abruptly and the row
        beside it stays put.
      </Text>
      <div className="flex items-center gap-3">
        <Button variant="primary" loading={loading} onClick={() => setLoading(true)}>
          Deploy component
        </Button>
        <Button variant="ghost" onClick={() => setLoading(false)}>
          Reset
        </Button>
      </div>
    </div>
  )
}

export const WithTooltip = () => (
  <div className="flex flex-wrap items-center gap-3 p-20">
    <Button variant="primary" disabled tooltip="Sync the app config first">
      Create install
    </Button>
    <Button variant="danger" disabled tooltip="You need admin access to delete this install">
      Delete install
    </Button>
    <Button variant="ghost" iconOnly aria-label="Close" tooltip="Close">
      <Icon variant="XIcon" size={16} />
    </Button>
    <Button variant="secondary" tooltip="Runs the deploy again from the last successful build" tooltipSide="bottom">
      Retry
    </Button>
  </div>
)

export const ModalFooter = () => (
  <div className="max-w-md p-8">
    <div className="flex flex-col gap-4 rounded-xl border border-divider bg-surface-01 p-5">
      <div className="flex items-start justify-between gap-4">
        <Text as="h2" variant="heading">
          Delete install?
        </Text>
        <Button variant="ghost" iconOnly aria-label="Close">
          <Icon variant="XIcon" size={16} />
        </Button>
      </div>
      <Text variant="body" color="secondary">
        This removes acme-production and everything it deployed. It cannot be
        undone.
      </Text>
      <div className="flex justify-end gap-2">
        <Button variant="ghost">Cancel</Button>
        <Button variant="danger">Delete install</Button>
      </div>
    </div>
  </div>
)

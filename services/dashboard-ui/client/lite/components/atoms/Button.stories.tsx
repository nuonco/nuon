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
    summary="Triggers an action. Four variants chosen by emphasis, not by colour."
    use={[
      'Anything that runs an action: submit, deploy, delete, open a modal.',
      'primary — the action the page exists for.',
      'secondary — the default, for every ordinary action.',
      'ghost — lowest emphasis: dismissals, toolbar actions, row actions.',
      'danger — entering a destructive flow.',
    ]}
    avoid={[
      'Navigation. A thing that changes the URL is a Link, even if it looks like a button.',
      'danger as the confirm button inside a destructive flow — that modal has already said what will happen, so its confirm is a primary.',
    ]}
    rules={[
      'At most one primary per page. Two primaries means neither is primary.',
      'iconOnly is a shape, not a variant — any variant can be icon-only, and it requires an aria-label.',
      'Every disabled button whose reason is not obvious from context gets a tooltip saying why. Pass tooltip; never hand-wrap Tooltip around a Button and never use title.',
      'Disabled uses aria-disabled, not the native attribute — native disabled swallows pointer events, so the tooltip explaining the disabled state would never show.',
      'A leading icon gets -ml-0.5 to cancel the transparent padding inside the SVG, so optical left and right insets both land at 14px.',
    ]}
    props={[
      { name: 'variant', type: "'primary' | 'secondary' | 'ghost' | 'danger'", default: "'secondary'", description: 'Emphasis level.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'Shows a spinner before the label, disables the button and sets aria-busy.' },
      { name: 'icon', type: 'ReactNode', description: 'Leading icon. Swapped out by the spinner while loading.' },
      { name: 'iconOnly', type: 'boolean', default: 'false', description: 'Square button. Icon goes in children; aria-label is required.' },
      { name: 'tooltip', type: 'ReactNode', description: 'Wraps the button in a Tooltip. The way to explain a disabled state.' },
      { name: 'tooltipSide', type: "'top' | 'bottom' | 'left' | 'right'", default: "'top'", description: 'Preferred tooltip side.' },
      { name: 'disabled', type: 'boolean', description: 'Renders aria-disabled, suppresses hover/active and blocks onClick, while staying focusable so its tooltip is reachable.' },
    ]}
    sections={[
      {
        heading: 'States',
        body: 'Hover, active, disabled and focus-visible are defined for every variant. Hover and active are suppressed while disabled, so a disabled button never lights up under the cursor. The focus ring is a 2px offset outline in --focus-ring and is never removed. Buttons set cursor-pointer explicitly, because Tailwind v4 Preflight no longer does.',
      },
      {
        heading: 'Loading',
        body: 'The spinner sits in a grid column that animates 0fr to 1fr, so the button grows into it over 200ms instead of snapping and shoving its neighbours. Height never changes. The transition is dropped under prefers-reduced-motion.',
      },
      {
        heading: 'Colour',
        body: 'Buttons have their own token family (--button-{variant}-bg/-hover/-active/-text) rather than reusing surface tokens, because the label sits on the button own background and has to clear contrast against that. Measured AA in light and dark, AAA in high contrast. The danger red is provisional — it is the one colour with no nuon.co equivalent and should fold into the status palette later.',
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

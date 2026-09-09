import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Status } from './Status'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Status',
}

const BY_THEME: Array<[string, string[]]> = [
  ['success', ['active', 'healthy', 'finished', 'approved']],
  ['error', ['failed', 'timed-out', 'unhealthy', 'policy-failed']],
  ['warn', ['degraded', 'drifted', 'cancelled', 'expired']],
  ['info', ['deploying', 'queued', 'provisioning', 'planning']],
  ['neutral', ['pending', 'not-deployed', 'disabled', 'unknown']],
  ['brand', ['special']],
]

export const Overview = () => (
  <ComponentDocs
    name="Status"
    tier="atom"
    summary="Renders an API status string as a themed chip, inline label, dot or icon."
    use={[
      'Render any status coming off the API, whether install, deploy, build, runner or component.',
      'Use chip in tables and lists, inline in prose or a detail row, dot where space is tight, and icon as a timeline or list marker.',
    ]}
    avoid={[
      'Do not use it for metadata that is not a state. That is a Badge.',
      'Avoid passing a theme by hand. Let the status string resolve it.',
    ]}
    rules={[
      'Pass a raw status string or the API status_v2 object; case and separators are normalised and an unknown status falls back to neutral.',
      'Every tooltip titles the status in bold with its icon; descriptions sit underneath when present.',
      'Tooltip descriptions wrap long tokens and visually clamp after ten lines.',
      'Icons resolve from the status independently of the theme; use the icon prop only when the surrounding context needs a different symbol.',
      'The theme prop is an escape hatch for statuses the shared map has not learned yet. Prefer adding the status to the map.',
      'A status is a filled chip and a badge is an outlined pill. Keep them structurally different so a state is never mistaken for metadata.',
      'Colour is never the only signal. Chip, inline and icon carry an icon; the dot names itself in a tooltip.',
    ]}
    props={[
      {
        name: 'status',
        type: 'string | TCompositeStatus',
        description: 'Raw API status or status_v2 object.',
      },
      {
        name: 'label',
        type: 'string',
        description: 'Overrides the humanised status text.',
      },
      {
        name: 'description',
        type: 'string',
        description:
          'Overrides the status_v2 description shown in the tooltip.',
      },
      {
        name: 'icon',
        type: 'TIconVariant',
        description:
          'Overrides the icon without changing the resolved status theme.',
      },
      {
        name: 'variant',
        type: "'chip' | 'inline' | 'dot' | 'icon'",
        default: "'chip'",
        description:
          'Filled chip, icon plus text, a named dot, or a tinted icon disc for timeline markers.',
      },
      {
        name: 'theme',
        type: "'success' | 'error' | 'warn' | 'info' | 'neutral' | 'brand'",
        description: 'Forces a theme instead of resolving from the status.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Shimmer while the status is unknown.',
      },
      {
        name: 'loadingWidth',
        type: 'number',
        description: 'Skeleton width in ch.',
      },
    ]}
    sections={[
      {
        heading: 'In-progress statuses',
        body: 'Statuses that mean work is in flight resolve to the info theme and show a spinner rather than a static icon, so a running deploy reads as moving.',
      },
    ]}
  />
)

export const Themes = () => (
  <div className="flex flex-col gap-6 p-8">
    {BY_THEME.map(([theme, statuses]) => (
      <div key={theme} className="flex flex-col gap-2">
        <Text variant="label" color="tertiary" family="mono">
          {theme}
        </Text>
        <div className="flex flex-wrap items-center gap-2">
          {statuses.map((status) => (
            <Status key={status} status={status} />
          ))}
        </div>
      </div>
    ))}
  </div>
)

export const Variants = () => (
  <div className="flex flex-col gap-6 p-8">
    {(['chip', 'inline', 'dot', 'icon'] as const).map((variant) => (
      <div key={variant} className="flex flex-col gap-2">
        <Text variant="label" color="tertiary" family="mono">
          {variant}
        </Text>
        <div className="flex flex-wrap items-center gap-4">
          <Status variant={variant} status="active" />
          <Status variant={variant} status="deploying" />
          <Status variant={variant} status="failed" />
          <Status variant={variant} status="degraded" />
          <Status variant={variant} status="pending" />
        </div>
      </div>
    ))}
  </div>
)

export const WithDescription = () => (
  <div className="flex flex-col items-start gap-6 p-8">
    {(['chip', 'inline', 'dot', 'icon'] as const).map((variant) => (
      <Status
        key={variant}
        variant={variant}
        status={{
          status: 'error',
          status_human_description:
            'Terraform apply exited with status 1 after the execution role was denied permission to create the requested resource.',
        }}
      />
    ))}
    <Status
      variant="dot"
      status={{
        status: 'error',
        status_human_description: `${'AccessDeniedException'.repeat(12)} ${'Unable to assume the execution role. '.repeat(12)}`,
      }}
    />
  </div>
)

export const IconOverride = () => (
  <div className="flex flex-col items-start gap-4 p-8">
    <Status variant="icon" status="pending" />
    <Status variant="icon" status="pending" icon="ProhibitIcon" />
    <Status status="pending" icon="ProhibitIcon" />
  </div>
)

export const UnknownStatus = () => (
  <div className="flex flex-col gap-3 p-8">
    <Status status="something-the-map-has-never-seen" />
    <Text variant="caption" color="tertiary">
      Falls back to neutral and humanises the string, rather than throwing or
      rendering nothing.
    </Text>
  </div>
)

export const InATableRow = () => (
  <div className="max-w-lg p-8">
    <div className="flex flex-col divide-y divide-divider rounded-xl border border-divider">
      {[
        ['acme-production', 'active'],
        ['acme-staging', 'deploying'],
        ['payments-eu', 'degraded'],
        ['payments-us', 'failed'],
      ].map(([name, status]) => (
        <div
          key={name}
          className="flex items-center justify-between gap-4 px-4 py-3"
        >
          <Text variant="body" family="mono">
            {name}
          </Text>
          <Status status={status} />
        </div>
      ))}
    </div>
  </div>
)

export const Loading = () => (
  <div className="flex flex-col gap-3 p-8">
    <div className="flex items-center gap-3">
      <Status status="active" loading />
      <Status status="active" />
    </div>
    <div className="flex items-center gap-3">
      <Status variant="icon" loading />
      <Status variant="icon" status="active" />
    </div>
  </div>
)

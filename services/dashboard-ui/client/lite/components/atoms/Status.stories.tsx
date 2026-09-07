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
    summary="Renders an API status string as a themed chip, inline label or dot."
    use={[
      'Render any status coming off the API, whether install, deploy, build, runner or component.',
      'Use chip in tables and lists, inline in prose or a detail row, and dot where space is tight.',
    ]}
    avoid={[
      'Do not use it for metadata that is not a state. That is a Badge.',
      'Avoid passing a theme by hand. Let the status string resolve it.',
    ]}
    rules={[
      'Pass the raw status string; case and separators are normalised and an unknown status falls back to neutral.',
      'The theme prop is an escape hatch for statuses the shared map has not learned yet. Prefer adding the status to the map.',
      'A status is a filled chip and a badge is an outlined pill. Keep them structurally different so a state is never mistaken for metadata.',
      'Colour is never the only signal. Every variant carries an icon or screen-reader text.',
    ]}
    props={[
      { name: 'status', type: 'string', description: 'Raw API status. Case and separators are normalised.' },
      { name: 'label', type: 'string', description: 'Overrides the humanised status text.' },
      { name: 'variant', type: "'chip' | 'inline' | 'dot'", default: "'chip'", description: 'Filled chip, icon plus text, or a bare dot with screen-reader text.' },
      { name: 'theme', type: "'success' | 'error' | 'warn' | 'info' | 'neutral' | 'brand'", description: 'Forces a theme instead of resolving from the status.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'Shimmer while the status is unknown.' },
      { name: 'loadingWidth', type: 'number', description: 'Skeleton width in ch.' },
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
    {(['chip', 'inline', 'dot'] as const).map((variant) => (
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
        <div key={name} className="flex items-center justify-between gap-4 px-4 py-3">
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
  <div className="flex items-center gap-3 p-8">
    <Status status="active" loading />
    <Status status="active" />
  </div>
)

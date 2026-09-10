import { DateTime } from 'luxon'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { Time, type TTimeFormat } from './Time'

export default {
  title: 'lite/molecules/Time',
}

const EXAMPLE = '2026-09-05T21:58:12.345-04:00'
const FORMATS: TTimeFormat[] = [
  'relative',
  'datetime',
  'full',
  'date',
  'time',
  'precise',
]

const Row = ({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) => (
  <div className="flex flex-wrap items-center justify-between gap-3">
    <Text variant="label" color="tertiary">
      {label}
    </Text>
    {children}
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="Time"
    tier="molecule"
    summary="A safely formatted instant with semantic markup, loading, and full timestamp context."
    use={[
      'Render API timestamps in tables, timelines, panels, metadata, and logs.',
      'Use relative for recent activity and datetime for stable metadata.',
      'Use precise only where milliseconds and the numeric offset matter.',
    ]}
    avoid={[
      'Do not format dates with Date or string slicing at the call site.',
      'Do not omit value to mean now; missing data should remain visibly missing.',
      'Do not disable the tooltip unless the complementary context is already adjacent.',
    ]}
    rules={[
      'String values are parsed as ISO timestamps; number values are Unix seconds.',
      'Every valid value renders a semantic time element with dateTime.',
      'The tooltip shows complementary context: full timestamp for short formats, relative time for full and precise.',
      'Relative values update every 30 seconds unless live is false.',
      'Missing and invalid values render the fallback instead of throwing or using now.',
    ]}
    props={[
      {
        name: 'value',
        type: 'string | number',
        description: 'ISO timestamp or Unix seconds.',
      },
      {
        name: 'format',
        type: "'relative' | 'datetime' | 'full' | 'date' | 'time' | 'precise'",
        default: "'datetime'",
        description: 'Display format for the instant.',
      },
      {
        name: 'tooltip',
        type: 'boolean',
        default: 'true',
        description:
          'Shows complementary context: full timestamp, or relative time when the display is already full or precise.',
      },
      {
        name: 'tooltipSide',
        type: "'top' | 'bottom' | 'left' | 'right'",
        default: "'top'",
        description:
          'Preferred tooltip side; flips automatically when it will not fit.',
      },
      {
        name: 'live',
        type: 'boolean',
        default: 'true for relative',
        description: 'Keeps relative text current.',
      },
      {
        name: 'fallback',
        type: 'ReactNode',
        default: "'—'",
        description: 'Content shown for missing or invalid values.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Uses the Text skeleton at the selected type size.',
      },
    ]}
  />
)

export const Formats = () => (
  <Card className="m-8 flex max-w-xl flex-col gap-4">
    {FORMATS.map((format) => (
      <Row key={format} label={format}>
        <Time value={EXAMPLE} format={format} />
      </Row>
    ))}
  </Card>
)

export const Relative = () => {
  const now = DateTime.now()
  return (
    <Card className="m-8 flex max-w-xl flex-col gap-4">
      <Row label="Just now">
        <Time value={now.minus({ seconds: 4 }).toISO()} format="relative" />
      </Row>
      <Row label="Past">
        <Time value={now.minus({ hours: 3 }).toISO()} format="relative" />
      </Row>
      <Row label="Future">
        <Time value={now.plus({ days: 2 }).toISO()} format="relative" />
      </Row>
    </Card>
  )
}

export const UnixSeconds = () => (
  <Card className="m-8 flex max-w-xl flex-col gap-4">
    <Row label="ISO string">
      <Time value="2022-01-01T00:00:00Z" />
    </Row>
    <Row label="Unix seconds">
      <Time value={1640995200} />
    </Row>
  </Card>
)

export const Loading = () => (
  <div className="flex flex-col gap-3 p-8">
    <Time loading />
    <Time loading variant="body" loadingWidth={24} />
    <Time loading variant="label" loadingWidth={12} />
  </div>
)

export const MissingAndInvalid = () => (
  <Card className="m-8 flex max-w-xl flex-col gap-4">
    <Row label="Missing">
      <Time />
    </Row>
    <Row label="Invalid">
      <Time value="not-a-timestamp" />
    </Row>
    <Row label="Custom fallback">
      <Time value="" fallback="Not recorded" />
    </Row>
  </Card>
)

export const WithoutTooltip = () => (
  <div className="p-8">
    <Time value={EXAMPLE} format="full" tooltip={false} />
  </div>
)

import { DateTime } from 'luxon'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { Duration, type TDurationFormat } from './Duration'

export default {
  title: 'lite/molecules/Duration',
}

const FORMATS: TDurationFormat[] = ['compact', 'long', 'timer']

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
    name="Duration"
    tier="molecule"
    summary="A readable elapsed time from nanoseconds or an ISO start/end range."
    use={[
      'Render API execution times from nanoseconds.',
      'Calculate elapsed time between start and end timestamps.',
      'Use timer for fixed-width technical displays.',
    ]}
    avoid={[
      'Do not calculate milliseconds or concatenate units at the call site.',
      'Do not use a timestamp component to display elapsed time.',
      'Do not treat zero as missing; zero is a valid duration.',
    ]}
    rules={[
      'nanoseconds takes precedence when both input forms are present.',
      'Compact output shows at most two meaningful units; long shows at most three.',
      'The tooltip preserves the complete human-readable duration when display is condensed.',
      'A start without an end updates every second unless live is false.',
      'Negative, missing, and invalid durations render the fallback.',
    ]}
    props={[
      {
        name: 'nanoseconds',
        type: 'number',
        description: 'Elapsed duration from the API.',
      },
      {
        name: 'start',
        type: 'string',
        description: 'ISO start timestamp.',
      },
      {
        name: 'end',
        type: 'string',
        description: 'ISO end timestamp; omitted means now.',
      },
      {
        name: 'format',
        type: "'compact' | 'long' | 'timer'",
        default: "'compact'",
        description: 'Readable, expanded, or clock-style output.',
      },
      {
        name: 'tooltip',
        type: 'boolean',
        default: 'true',
        description: 'Shows the complete duration when display is condensed.',
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
        default: 'true for an open range',
        description: 'Keeps an unfinished range current.',
      },
      {
        name: 'fallback',
        type: 'ReactNode',
        default: "'—'",
        description: 'Content shown for unusable input.',
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
        <Duration nanoseconds={93_784_000_000_000} format={format} />
      </Row>
    ))}
  </Card>
)

export const Ranges = () => (
  <Card className="m-8 flex max-w-xl flex-col gap-4">
    <Row label="Finished">
      <Duration start="2026-09-05T20:00:00Z" end="2026-09-05T21:32:18Z" />
    </Row>
    <Row label="In progress">
      <Duration start={DateTime.now().minus({ minutes: 8 }).toISO()} />
    </Row>
  </Card>
)

export const Scales = () => (
  <Card className="m-8 flex max-w-xl flex-col gap-4">
    <Row label="Sub-millisecond">
      <Duration nanoseconds={500_000} />
    </Row>
    <Row label="Milliseconds">
      <Duration nanoseconds={250_000_000} />
    </Row>
    <Row label="Seconds">
      <Duration nanoseconds={45_000_000_000} />
    </Row>
    <Row label="Minutes">
      <Duration nanoseconds={325_000_000_000} />
    </Row>
    <Row label="Hours">
      <Duration nanoseconds={9_845_000_000_000} />
    </Row>
    <Row label="Days">
      <Duration nanoseconds={190_845_000_000_000} />
    </Row>
  </Card>
)

export const Zero = () => (
  <div className="p-8">
    <Duration nanoseconds={0} />
  </div>
)

export const Loading = () => (
  <div className="flex flex-col gap-3 p-8">
    <Duration loading />
    <Duration loading variant="body" loadingWidth={12} />
  </div>
)

export const MissingAndInvalid = () => (
  <Card className="m-8 flex max-w-xl flex-col gap-4">
    <Row label="Missing">
      <Duration />
    </Row>
    <Row label="Invalid range">
      <Duration start="not-a-timestamp" end="also-invalid" />
    </Row>
    <Row label="Negative">
      <Duration nanoseconds={-10} />
    </Row>
    <Row label="Custom fallback">
      <Duration fallback="Not recorded" />
    </Row>
  </Card>
)

import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { Cron, type TCronFormat } from './Cron'

export default {
  title: 'lite/molecules/Cron',
}

const FORMATS: TCronFormat[] = ['human', 'expression', 'both']

const Row = ({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) => (
  <div className="flex flex-wrap items-start justify-between gap-3">
    <Text variant="label" color="tertiary">
      {label}
    </Text>
    {children}
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="Cron"
    tier="molecule"
    summary="A cron schedule rendered as plain language, its expression, or both."
    use={[
      'Explain drift detection, maintenance, and other recurring schedules.',
      'Use human for product UI and expression for configuration detail.',
      'Use both when users need to compare intent with stored configuration.',
    ]}
    avoid={[
      'Do not use it to validate form input; schema validation owns that.',
      'Do not hand-write descriptions for known cron expressions.',
      'Do not hide an invalid expression behind a plausible description.',
    ]}
    rules={[
      'Human display reveals the raw expression in a tooltip by default.',
      'Expression display reveals the human meaning in a tooltip.',
      'Invalid expressions display Invalid schedule and preserve the raw value in the tooltip.',
      'Missing expressions render the fallback.',
    ]}
    props={[
      {
        name: 'value',
        type: 'string',
        description: 'Cron expression to describe.',
      },
      {
        name: 'format',
        type: "'human' | 'expression' | 'both'",
        default: "'human'",
        description: 'Plain-language, raw, or combined presentation.',
      },
      {
        name: 'tooltip',
        type: 'boolean',
        default: 'true',
        description: 'Shows the alternate representation.',
      },
      {
        name: 'fallback',
        type: 'ReactNode',
        default: "'—'",
        description: 'Content shown when no expression is present.',
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
        <Cron value="0 */6 * * *" format={format} />
      </Row>
    ))}
  </Card>
)

export const CommonSchedules = () => (
  <Card className="m-8 flex max-w-xl flex-col gap-4">
    <Row label="Every 15 minutes">
      <Cron value="*/15 * * * *" />
    </Row>
    <Row label="Daily">
      <Cron value="0 2 * * *" />
    </Row>
    <Row label="Weekdays">
      <Cron value="0 9 * * 1-5" />
    </Row>
    <Row label="Monthly">
      <Cron value="0 0 1 * *" />
    </Row>
  </Card>
)

export const Expression = () => (
  <div className="p-8">
    <Cron value="0 9 * * 1-5" format="expression" />
  </div>
)

export const Loading = () => (
  <div className="flex flex-col gap-3 p-8">
    <Cron loading />
    <Cron loading variant="body" loadingWidth={24} />
  </div>
)

export const MissingAndInvalid = () => (
  <Card className="m-8 flex max-w-xl flex-col gap-4">
    <Row label="Missing">
      <Cron />
    </Row>
    <Row label="Invalid">
      <Cron value="not a cron expression" />
    </Row>
    <Row label="Custom fallback">
      <Cron value="" fallback="Not scheduled" />
    </Row>
  </Card>
)

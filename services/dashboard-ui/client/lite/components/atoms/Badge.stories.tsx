import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Badge } from './Badge'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Badge',
}

const LABEL_COLORS = ['#4cc9f0', '#f72585', '#3a00ff', '#4aa578', '#e8a33d', '#8b5cf6']

export const Overview = () => (
  <ComponentDocs
    name="Badge"
    tier="atom"
    summary="A small piece of metadata, as a single pill or a key/value label pair."
    use={[
      'Show metadata about a resource, such as a label, a count or a short classification.',
      'Pass labelKey with labelValue for a key/value label, which renders as one joined pill.',
      'Pass color for a user-chosen label colour, taken straight from the API.',
    ]}
    avoid={[
      'Do not use a badge for resource state. That is Status.',
      'Do not make a badge clickable. Removal is the one exception, and it renders a real button inside.',
    ]}
    rules={[
      'Pass labelKey and labelValue rather than composing two Badges.',
      'A user-chosen colour applies to the value half only, and is ignored in the high contrast theme.',
      'Long keys and values truncate to one line.',
    ]}
    props={[
      { name: 'tone', type: "'neutral' | 'accent'", default: "'neutral'", description: 'Semantic tone.' },
      { name: 'variant', type: "'default' | 'code'", default: "'default'", description: 'The default is sans and pill-shaped; code is mono with a softer radius.' },
      { name: 'color', type: 'string', description: 'User-chosen label colour, any CSS colour. Applies to the value half.' },
      { name: 'labelKey', type: 'string', description: 'Left half. Its presence switches to the key/value shape.' },
      { name: 'labelValue', type: 'string', description: 'Right half, the coloured one.' },
      { name: 'onRemove', type: '() => void', description: 'Renders a remove button inside the badge.' },
      { name: 'removeLabel', type: 'string', default: "'Remove'", description: 'Accessible name for the remove button.' },
      { name: 'disabled', type: 'boolean', default: 'false', description: 'Disables the remove button.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'Shimmer in the badge shape.' },
      { name: 'loadingWidth', type: 'number', description: 'Skeleton width in ch.' },
    ]}
  />
)

export const Tones = () => (
  <div className="flex flex-wrap items-center gap-2 p-8">
    <Badge tone="neutral">neutral</Badge>
    <Badge tone="accent">accent</Badge>
  </div>
)

export const Variants = () => (
  <div className="flex flex-col gap-4 p-8">
    <div className="flex items-center gap-2">
      <Badge variant="default">production</Badge>
      <Text variant="caption" color="tertiary">
        default — sans, pill
      </Text>
    </div>
    <div className="flex items-center gap-2">
      <Badge variant="code">inst_01h9k2m4p6</Badge>
      <Text variant="caption" color="tertiary">
        code — mono, softer radius
      </Text>
    </div>
  </div>
)

export const Labels = () => (
  <div className="flex flex-wrap items-center gap-2 p-8">
    <Badge variant="code" labelKey="env" labelValue="production" />
    <Badge variant="code" labelKey="region" labelValue="us-west-2" />
    <Badge variant="code" labelKey="tier" labelValue="enterprise" />
  </div>
)

export const CustomColors = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      One hex from the API, mixed per theme. Switch the Ladle theme to see both,
      and to high contrast to see the colour deliberately dropped.
    </Text>
    <div className="flex flex-wrap items-center gap-2">
      {LABEL_COLORS.map((color) => (
        <Badge key={color} variant="code" labelKey="team" labelValue={color} color={color} />
      ))}
    </div>
  </div>
)

export const Removable = () => (
  <div className="flex flex-wrap items-center gap-2 p-8">
    <Badge variant="code" labelKey="env" labelValue="production" onRemove={() => {}} />
    <Badge variant="code" labelKey="team" labelValue="payments" color="#4cc9f0" onRemove={() => {}} />
    <Badge onRemove={() => {}}>dismissable</Badge>
    <Badge variant="code" labelKey="env" labelValue="locked" onRemove={() => {}} disabled />
  </div>
)

export const Truncated = () => (
  <div className="max-w-xs p-8">
    <Badge
      variant="code"
      labelKey="a-very-long-label-key-that-keeps-going"
      labelValue="and-an-even-longer-value-that-should-clamp-rather-than-overflow"
    />
  </div>
)

export const Loading = () => (
  <div className="flex items-center gap-2 p-8">
    <Badge loading />
    <Badge loading variant="code" loadingWidth={14} />
    <Badge>loaded</Badge>
  </div>
)

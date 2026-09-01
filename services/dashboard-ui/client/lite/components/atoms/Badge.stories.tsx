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
    summary="A small piece of metadata. Renders either a single pill or a key/value label pair, including labels whose colour the user picked."
    use={[
      'Metadata about a resource: a label, a count, a short classification.',
      'labelKey + labelValue for a key/value label, which renders as one joined pill.',
      'color for a user-chosen label colour, taken straight from the API.',
    ]}
    avoid={[
      'Resource status — that is Status, once the status palette exists. A badge that means "failed" should not be a hand-toned Badge.',
      'Anything clickable. A badge is not a button; onRemove is the one exception and it is a real button inside.',
    ]}
    rules={[
      'One size. The old dashboard had four (xs/sm/md/lg) plus a second table of hand-measured skeleton heights for each.',
      'Two tones only. A positive/success tone needs its own tinted surface, not a recolour of the accent tint — it waits for the status palette.',
      'The key/value pair is one component, not two Badges glued with rounded-r-none. Callers pass labelKey and labelValue.',
      'A user-chosen colour is mixed against theme anchors so it works in light and dark, and is ignored entirely in high contrast, where an arbitrary hex cannot be made to clear AAA.',
    ]}
    props={[
      { name: 'tone', type: "'neutral' | 'accent'", default: "'neutral'", description: 'Semantic tone. Anything meaning success/failure/pending waits for Status and the status palette.' },
      { name: 'variant', type: "'default' | 'code'", default: "'default'", description: 'default is sans and pill-shaped; code is mono with a softer radius.' },
      { name: 'color', type: 'string', description: 'User-chosen label colour, any CSS colour. Applies to the value half.' },
      { name: 'labelKey', type: 'string', description: 'Left half. Presence of this switches to the key/value shape.' },
      { name: 'labelValue', type: 'string', description: 'Right half, the coloured one.' },
      { name: 'onRemove', type: '() => void', description: 'Renders a remove button inside the badge.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'Shimmer in the badge shape.' },
    ]}
    sections={[
      {
        heading: 'Custom colour and contrast',
        body: 'The value half mixes the chosen colour against a surface anchor for the background and against black or white for the text, per theme, so one hex from the API reads in both light and dark. It is an approximation, not a guarantee — an arbitrary user colour cannot be proven to clear AA, which is exactly why high contrast drops the colour and falls back to the neutral treatment.',
      },
      {
        heading: 'Truncation',
        body: 'Long keys and values clamp to one line. The old LabelBadge measured scrollWidth against clientWidth on every resize to decide whether to attach a tooltip; lite will hand that to Tooltip when it exists rather than running a resize listener per badge.',
      },
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

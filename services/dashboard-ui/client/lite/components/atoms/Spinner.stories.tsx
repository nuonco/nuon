import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Button } from './Button'
import { Spinner } from './Spinner'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Spinner',
}

export const Overview = () => (
  <ComponentDocs
    name="Spinner"
    tier="atom"
    summary="An indeterminate progress indicator for work already in flight, drawn as a track plus a rotating arc."
    use={[
      'Inside a Button while its action is in flight — Button does this for you via loading.',
      'Any short wait where the shape of the incoming content is unknown.',
    ]}
    avoid={[
      'Waiting on content whose shape you already know — that is a skeleton, via the loading prop on Text, Link or Badge.',
      'Long or multi-step work, where progress or a status message tells the user more than a spinner does.',
    ]}
    rules={[
      'currentColor, so it inherits the colour of whatever it sits in. It never takes a colour prop.',
      'Not an Icon. Phosphor glyphs sit in a 256 viewBox with about 12.5% transparent padding a side, so a rotating one wobbles off-centre; this is drawn on its own 16 viewBox centred on the rotation axis.',
      'Decorative by default (aria-hidden). Pass label only when the spinner is the sole announcement of the wait — inside a Button the button already owns aria-busy.',
      'Stops animating under prefers-reduced-motion, leaving the static ring.',
    ]}
    props={[
      {
        name: 'size',
        type: 'number',
        default: '16',
        description: 'Width and height in px. Stroke stays proportional.',
      },
      {
        name: 'label',
        type: 'string',
        description:
          'Makes it a live status with this accessible name. Omit when something else announces the wait.',
      },
    ]}
  />
)

export const Sizes = () => (
  <div className="flex items-center gap-6 p-8">
    {[12, 16, 20, 24, 32, 48].map((size) => (
      <div key={size} className="flex flex-col items-center gap-2">
        <Spinner size={size} />
        <Text variant="label" color="tertiary" family="mono">
          {size}
        </Text>
      </div>
    ))}
  </div>
)

export const InheritsColor = () => (
  <div className="flex items-center gap-6 p-8">
    {(['primary', 'secondary', 'tertiary', 'accent'] as const).map((color) => (
      <div key={color} className="flex flex-col items-center gap-2">
        <Text color={color}>
          <Spinner size={24} />
        </Text>
        <Text variant="label" color="tertiary" family="mono">
          {color}
        </Text>
      </div>
    ))}
  </div>
)

export const InAButton = () => (
  <div className="flex flex-wrap items-center gap-3 p-8">
    <Button variant="primary" loading>
      Deploying component
    </Button>
    <Button variant="secondary" loading>
      Retrying
    </Button>
    <Button variant="danger" loading>
      Deleting install
    </Button>
  </div>
)

export const WithLabel = () => (
  <div className="flex items-center gap-2 p-8">
    <Spinner size={20} label="Loading installs" />
    <Text variant="caption" color="tertiary">
      Announced as a status. Only do this when nothing else reports the wait.
    </Text>
  </div>
)

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
    summary="An indeterminate progress indicator for work already in flight."
    use={[
      'Show one inside a button while its action runs, which the button does for you via loading.',
      'Cover a short wait where the shape of the incoming content is unknown.',
    ]}
    avoid={[
      'Do not use one for content whose shape you already know. That is a skeleton, via the loading prop on Text, Link, Badge or ID.',
      'Avoid it for long or multi-step work, where progress or a status message says more.',
    ]}
    rules={[
      'The spinner uses the current text colour, so it inherits whatever it sits in.',
      'The spinner is decorative by default. Pass a label only when it is the sole announcement of the wait, since a button already reports its own.',
    ]}
    props={[
      { name: 'size', type: 'number', default: '16', description: 'Width and height in px.' },
      { name: 'label', type: 'string', description: 'Makes it a live status with this accessible name.' },
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

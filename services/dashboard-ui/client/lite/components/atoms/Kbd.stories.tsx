import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Button } from './Button'
import { Icon } from './Icon'
import { Kbd } from './Kbd'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Kbd',
}

export const Overview = () => (
  <ComponentDocs
    name="Kbd"
    tier="atom"
    summary="A key cap for naming the key that performs an action."
    use={[
      'Name the shortcut inside a tooltip on the control it drives.',
      'Show an arrow key as its icon rather than the word Up or Down.',
    ]}
    avoid={[
      'Do not use it as a badge for arbitrary short text.',
      'Do not spell a shortcut in prose when the control already has a tooltip.',
    ]}
    rules={[
      'It renders a real kbd element, so a screen reader announces it as a key.',
      'An arrow key is the icon; a named key such as Enter or Esc is its word.',
      'It sizes itself for one key. Render one cap per key and separate them with a thin plus.',
    ]}
    props={[
      {
        name: 'children',
        type: 'ReactNode',
        description: 'The key: an Icon for arrows, text for anything named.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="flex flex-wrap items-center gap-3 p-8">
    <Kbd>
      <Icon variant="ArrowUpIcon" size={10} />
    </Kbd>
    <Kbd>
      <Icon variant="ArrowDownIcon" size={10} />
    </Kbd>
    <Kbd>Enter</Kbd>
    <Kbd>Esc</Kbd>
    <Kbd>Space</Kbd>
  </div>
)

export const InSentence = () => (
  <div className="max-w-md p-8">
    <Text variant="caption" color="tertiary">
      Press <Kbd>Enter</Kbd> to jump to the next match, or{' '}
      <Kbd>
        <Icon variant="ArrowUpIcon" size={10} />
      </Kbd>{' '}
      to step back.
    </Text>
  </div>
)

export const Combination = () => (
  <div className="flex items-center gap-1 p-8">
    <Kbd>Shift</Kbd>
    <Text variant="caption" color="tertiary">
      +
    </Text>
    <Kbd>Enter</Kbd>
  </div>
)

export const InTooltip = () => (
  <div className="flex items-center gap-2 p-20">
    <Button
      size="sm"
      variant="ghost"
      iconOnly
      aria-label="Previous match"
      tooltip={
        <span className="flex items-center gap-1.5">
          <Text variant="caption">Previous match</Text>
          <Kbd>
            <Icon variant="ArrowUpIcon" size={10} />
          </Kbd>
        </span>
      }
    >
      <Icon variant="CaretUpIcon" size={14} />
    </Button>
    <Button
      size="sm"
      variant="ghost"
      iconOnly
      aria-label="Next match"
      tooltip={
        <span className="flex items-center gap-1.5">
          <Text variant="caption">Next match</Text>
          <Kbd>
            <Icon variant="ArrowDownIcon" size={10} />
          </Kbd>
        </span>
      }
    >
      <Icon variant="CaretDownIcon" size={14} />
    </Button>
  </div>
)

import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Icon } from './Icon'
import { ToggleButton } from './ToggleButton'

export default {
  title: 'lite/atoms/ToggleButton',
}

export const Overview = () => (
  <ComponentDocs
    name="ToggleButton"
    tier="atom"
    summary="A controlled group of two or more buttons for choosing one view or mode."
    use={[
      'Switch between equivalent presentations or modes.',
      'Use concise text labels or familiar icons with accessible names.',
    ]}
    avoid={[
      'Do not use it for independent options that can be active together.',
      'Do not use it for navigation or values submitted as form data.',
    ]}
    rules={[
      'Exactly one option is pressed at a time.',
      'Every icon-only option has an accessible label and tooltip.',
      'The group always has an accessible label.',
    ]}
    props={[
      {
        name: 'value',
        type: 'TValue',
        description: 'The currently selected option value.',
      },
      {
        name: 'options',
        type: '[Option<TValue>, Option<TValue>, ...Option<TValue>[]]',
        description: 'Two or more choices rendered in their declared order.',
      },
      {
        name: 'onValueChange',
        type: '(value: TValue) => void',
        description: 'Receives the selected option value.',
      },
      {
        name: 'label',
        type: 'string',
        description: 'Accessible name for the button group.',
      },
      {
        name: 'size',
        type: "'sm' | 'md'",
        default: "'sm'",
        description: 'Button size shared by every option.',
      },
    ]}
  />
)

export const TextLabels = () => {
  const [value, setValue] = useState<'day' | 'week' | 'month'>('week')

  return (
    <div className="p-8">
      <ToggleButton
        label="Time range"
        value={value}
        onValueChange={setValue}
        options={[
          { value: 'day', label: 'Day' },
          { value: 'week', label: 'Week' },
          { value: 'month', label: 'Month' },
        ]}
      />
    </div>
  )
}

export const IconOnly = () => {
  const [value, setValue] = useState<'table' | 'cards'>('table')

  return (
    <div className="p-8">
      <ToggleButton
        label="Collection view"
        value={value}
        onValueChange={setValue}
        options={[
          {
            value: 'table',
            label: <Icon variant="TableIcon" size={14} aria-hidden />,
            ariaLabel: 'Table view',
            tooltip: 'Table view',
            iconOnly: true,
          },
          {
            value: 'cards',
            label: <Icon variant="SquaresFourIcon" size={14} aria-hidden />,
            ariaLabel: 'Card view',
            tooltip: 'Card view',
            iconOnly: true,
          },
        ]}
      />
    </div>
  )
}

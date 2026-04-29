export default {
  title: 'Common/ToggleButton',
}

import { useState } from 'react'
import { Icon } from './Icon'
import { ToggleButton } from './ToggleButton'

export const Default = () => {
  const [value, setValue] = useState('grid')
  return (
    <ToggleButton
      value={value}
      onChange={setValue}
      options={[
        { value: 'grid', label: <Icon variant="ListDashes" size={16} />, ariaLabel: 'Grid view' },
        { value: 'json', label: <Icon variant="BracketsCurly" size={16} />, ariaLabel: 'JSON view' },
      ]}
    />
  )
}

export const ThreeOptions = () => {
  const [value, setValue] = useState('list')
  return (
    <ToggleButton
      value={value}
      onChange={setValue}
      options={[
        { value: 'list', label: <Icon variant="List" size={16} />, ariaLabel: 'List view' },
        { value: 'grid', label: <Icon variant="GridFour" size={16} />, ariaLabel: 'Grid view' },
        { value: 'table', label: <Icon variant="Table" size={16} />, ariaLabel: 'Table view' },
      ]}
    />
  )
}

export const WithTextLabels = () => {
  const [value, setValue] = useState('daily')
  return (
    <ToggleButton
      value={value}
      onChange={setValue}
      options={[
        { value: 'daily', label: 'Daily' },
        { value: 'weekly', label: 'Weekly' },
        { value: 'monthly', label: 'Monthly' },
      ]}
    />
  )
}

export const MediumSize = () => {
  const [value, setValue] = useState('left')
  return (
    <ToggleButton
      size="md"
      value={value}
      onChange={setValue}
      options={[
        { value: 'left', label: 'Left' },
        { value: 'right', label: 'Right' },
      ]}
    />
  )
}
